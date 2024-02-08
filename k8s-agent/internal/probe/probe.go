/*
Copyright 2023-2024 VMware Inc.
SPDX-License-Identifier: Apache-2.0

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package probe

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"

	K8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"collie-agent/internal/model"
	"collie-agent/internal/reporter"
)

type Probe struct {
	ctx       context.Context
	log       *logrus.Entry
	clientset *kubernetes.Clientset
	cc        *reporter.CollieClient
}

func New(ctx context.Context, log *logrus.Entry, clientset *kubernetes.Clientset, cc *reporter.CollieClient) *Probe {
	return &Probe{ctx, log, clientset, cc}
}

func GetClusterId(ctx context.Context, log *logrus.Entry, clientset *kubernetes.Clientset) (string, error) {

	// retrieve the cluster information
	nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", err
	}
	if len(nodes.Items) == 0 {
		return "", fmt.Errorf("no nodes found in the cluster")
	}

	// extract the cluster ID from the node information
	clusterID := nodes.Items[0].Status.NodeInfo.SystemUUID
	log.Infoln("clusterId: ", clusterID)
	return clusterID, nil
}

func (p *Probe) DiscoverCluster() error {

	log := p.log

	log.Info("DiscoverCluster start")

	startTime := time.Now()
	defer func() {
		p.cc.DeleteOldDoc(startTime, "cluster")
		log.Info("DiscoverCluster exit")
	}()

	serverVersion, err := p.clientset.Discovery().ServerVersion()
	if err != nil {
		return err
	}

	serverVersionJson, err := json.Marshal(serverVersion)
	if err != nil {
		return err
	}
	log.Printf("Cluster info: %s", string(serverVersionJson))

	info := model.ClusterInfo{
		Provider: "",
		Data:     serverVersion,
	}

	p.cc.ReportClusterInfo(info)
	return nil
}

func (p *Probe) DiscoverResources() error {

	log := p.log

	log.Info("DiscoverResources start")
	startTime := time.Now()
	defer func() {
		p.cc.DeleteOldDoc(startTime, "resource")
		log.Info("DiscoverResources exit")
	}()

	coreV1 := p.clientset.CoreV1()

	namespaceList, err := coreV1.Namespaces().List(p.ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	//ignoredNamespaces := map[string]int{"kube-node-lease":1, "kube-public":1, "kube-system":1}
	ignoredNamespaces := map[string]int{}

	p.discoverNodes()
	p.discoverPersistentVolumes()
	p.discoverCSIDrivers()
	p.discoverCSINodes()
	p.discoverStorageClasses()

	for _, ns := range namespaceList.Items {
		namespace := ns.Name
		log.Infoln("namespace", namespace)
		if _, ok := ignoredNamespaces[namespace]; ok {
			log.Info("Ignore namespace ", namespace)
			continue
		}

		p.discoverPods(namespace)
		p.discoverReplicationControllers(namespace)
		p.discoverCSIStorageCapacities(namespace)
		p.discoverJobs(namespace)
		p.discoverCronJobs(namespace)
		p.discoverDaemonSets(namespace)
		p.discoverDeployments(namespace)
		p.discoverServices(namespace)
		p.discoverStatefulSets(namespace)
		p.discoverEvents(namespace)
		p.discoverHorizontalPodAutoscalers(namespace)
		p.discoverLeases(namespace)
		p.discoverPersistentVolumeClaims(namespace)
		p.discoverReplicaSets(namespace)
	}
	return nil
}

func (p *Probe) reportResource(name string, data interface{}, err error) {
	if K8sErrors.IsNotFound(err) {
		p.log.Printf("Resource not found: %s", name)
	} else if statusError, isStatus := err.(*K8sErrors.StatusError); isStatus {
		p.log.Printf("Error getting resource: %s, %v", name, statusError.ErrStatus.Message)
	} else if err != nil {
		p.cc.ReportError("get-res", name, err)
	} else {
		p.cc.ReportResource(name, data)
	}
}

func (p *Probe) discoverNodes() {
	kind := "nodes"
	log := p.log
	log.Info("discoverRes start: ", kind)
	defer func() {
		log.Info("discoverRes exit: ", kind)
	}()

	coreV1 := p.clientset.CoreV1()
	api := coreV1.Nodes()
	itemList, err := api.List(p.ctx, metav1.ListOptions{})
	if err != nil {
		p.cc.ReportError("list-res", kind, err)
		return
	}
	total := len(itemList.Items)
	for idx, item := range itemList.Items {
		resourceName := kind + "#" + item.Name
		p.log.Printf("Resource %d/%d: %s", idx+1, total, resourceName)
		itemInfo, err := api.Get(p.ctx, item.Name, metav1.GetOptions{})
		p.reportResource(resourceName, itemInfo, err)
	}
}

func (p *Probe) discoverPersistentVolumes() {
	kind := "persistentvolumes"
	log := p.log
	log.Info("discoverRes start: ", kind)
	defer func() {
		log.Info("discoverRes exit: ", kind)
	}()

	coreV1 := p.clientset.CoreV1()
	api := coreV1.PersistentVolumes()
	itemList, err := api.List(p.ctx, metav1.ListOptions{})
	if err != nil {
		p.cc.ReportError("list-res", kind, err)
		return
	}
	total := len(itemList.Items)
	for idx, item := range itemList.Items {
		resourceName := kind + "#" + item.Name
		p.log.Printf("Resource %d/%d: %s", idx+1, total, resourceName)
		itemInfo, err := api.Get(p.ctx, item.Name, metav1.GetOptions{})
		p.reportResource(resourceName, itemInfo, err)
	}
}

func (p *Probe) discoverPods(namespace string) {
	kind := "pods"
	log := p.log
	log.Info("discoverRes start: ", kind)
	defer func() {
		log.Info("discoverRes exit: ", kind)
	}()

	coreV1 := p.clientset.CoreV1()
	api := coreV1.Pods(namespace)
	itemList, err := api.List(p.ctx, metav1.ListOptions{})
	if err != nil {
		p.cc.ReportError("list-res", kind+"#"+namespace, err)
		return
	}
	total := len(itemList.Items)
	for idx, item := range itemList.Items {
		resourceName := kind + "#" + namespace + "/" + item.Name
		p.log.Printf("Resource %d/%d: %s", idx+1, total, resourceName)
		itemInfo, err := api.Get(p.ctx, item.Name, metav1.GetOptions{})
		p.reportResource(resourceName, itemInfo, err)
	}
}

func (p *Probe) discoverReplicationControllers(namespace string) {
	kind := "replicationcontrollers"
	log := p.log
	log.Info("discoverRes start: ", kind)
	defer func() {
		log.Info("discoverRes exit: ", kind)
	}()

	coreV1 := p.clientset.CoreV1()
	api := coreV1.ReplicationControllers(namespace)
	itemList, err := api.List(p.ctx, metav1.ListOptions{})
	if err != nil {
		p.cc.ReportError("list-res", kind+"#"+namespace, err)
		return
	}
	total := len(itemList.Items)
	for idx, item := range itemList.Items {
		resourceName := kind + "#" + namespace + "/" + item.Name
		p.log.Printf("Resource %d/%d: %s", idx+1, total, resourceName)
		itemInfo, err := api.Get(p.ctx, item.Name, metav1.GetOptions{})
		p.reportResource(resourceName, itemInfo, err)
	}
}

func (p *Probe) discoverPersistentVolumeClaims(namespace string) {
	kind := "persistentvolumeclaims"
	log := p.log
	log.Info("discoverRes start: ", kind)
	defer func() {
		log.Info("discoverRes exit: ", kind)
	}()

	coreV1 := p.clientset.CoreV1()
	api := coreV1.PersistentVolumeClaims(namespace)
	itemList, err := api.List(p.ctx, metav1.ListOptions{})
	if err != nil {
		p.cc.ReportError("list-res", kind+"#"+namespace, err)
		return
	}
	total := len(itemList.Items)
	for idx, item := range itemList.Items {
		resourceName := kind + "#" + namespace + "/" + item.Name
		p.log.Printf("Resource %d/%d: %s", idx+1, total, resourceName)
		itemInfo, err := api.Get(p.ctx, item.Name, metav1.GetOptions{})
		p.reportResource(resourceName, itemInfo, err)
	}
}

func (p *Probe) discoverServices(namespace string) {
	kind := "services"
	log := p.log
	log.Info("discoverRes start: ", kind)
	defer func() {
		log.Info("discoverRes exit: ", kind)
	}()

	coreV1 := p.clientset.CoreV1()
	api := coreV1.Services(namespace)
	itemList, err := api.List(p.ctx, metav1.ListOptions{})
	if err != nil {
		p.cc.ReportError("list-res", kind+"#"+namespace, err)
		return
	}
	total := len(itemList.Items)
	for idx, item := range itemList.Items {
		resourceName := kind + "#" + namespace + "/" + item.Name
		p.log.Printf("Resource %d/%d: %s", idx+1, total, resourceName)
		itemInfo, err := api.Get(p.ctx, item.Name, metav1.GetOptions{})
		p.reportResource(resourceName, itemInfo, err)
	}
}

func (p *Probe) discoverEvents(namespace string) {
	kind := "events"
	log := p.log
	log.Info("discoverRes start: ", kind)
	defer func() {
		log.Info("discoverRes exit: ", kind)
	}()

	coreV1 := p.clientset.CoreV1()
	api := coreV1.Events(namespace)
	itemList, err := api.List(p.ctx, metav1.ListOptions{})
	if err != nil {
		p.cc.ReportError("list-res", kind+"#"+namespace, err)
		return
	}
	total := len(itemList.Items)
	for idx, item := range itemList.Items {
		resourceName := kind + "#" + namespace + "/" + item.Name
		p.log.Printf("Resource %d/%d: %s", idx+1, total, resourceName)
		itemInfo, err := api.Get(p.ctx, item.Name, metav1.GetOptions{})
		p.reportResource(resourceName, itemInfo, err)
	}
}

func (p *Probe) discoverDeployments(namespace string) {
	kind := "deployments"
	log := p.log
	log.Info("discoverRes start: ", kind)
	defer func() {
		log.Info("discoverRes exit: ", kind)
	}()

	appsV1 := p.clientset.AppsV1()
	api := appsV1.Deployments(namespace)
	itemList, err := api.List(p.ctx, metav1.ListOptions{})
	if err != nil {
		p.cc.ReportError("list-res", kind+"#"+namespace, err)
		return
	}
	total := len(itemList.Items)
	for idx, item := range itemList.Items {
		resourceName := kind + "#" + namespace + "/" + item.Name
		p.log.Printf("Resource %d/%d: %s", idx+1, total, resourceName)
		itemInfo, err := api.Get(p.ctx, item.Name, metav1.GetOptions{})
		p.reportResource(resourceName, itemInfo, err)
	}
}

func (p *Probe) discoverReplicaSets(namespace string) {
	kind := "replicasets"
	log := p.log
	log.Info("discoverRes start: ", kind)
	defer func() {
		log.Info("discoverRes exit: ", kind)
	}()

	appsV1 := p.clientset.AppsV1()
	api := appsV1.ReplicaSets(namespace)
	itemList, err := api.List(p.ctx, metav1.ListOptions{})
	if err != nil {
		p.cc.ReportError("list-res", kind+"#"+namespace, err)
		return
	}
	total := len(itemList.Items)
	for idx, item := range itemList.Items {
		resourceName := kind + "#" + namespace + "/" + item.Name
		p.log.Printf("Resource %d/%d: %s", idx+1, total, resourceName)
		itemInfo, err := api.Get(p.ctx, item.Name, metav1.GetOptions{})
		p.reportResource(resourceName, itemInfo, err)
	}
}

func (p *Probe) discoverDaemonSets(namespace string) {
	kind := "daemonsets"
	log := p.log
	log.Info("discoverRes start: ", kind)
	defer func() {
		log.Info("discoverRes exit: ", kind)
	}()

	appsV1 := p.clientset.AppsV1()
	api := appsV1.DaemonSets(namespace)
	itemList, err := api.List(p.ctx, metav1.ListOptions{})
	if err != nil {
		p.cc.ReportError("list-res", kind+"#"+namespace, err)
		return
	}
	total := len(itemList.Items)
	for idx, item := range itemList.Items {
		resourceName := kind + "#" + namespace + "/" + item.Name
		p.log.Printf("Resource %d/%d: %s", idx+1, total, resourceName)
		itemInfo, err := api.Get(p.ctx, item.Name, metav1.GetOptions{})
		p.reportResource(resourceName, itemInfo, err)
	}
}

func (p *Probe) discoverStatefulSets(namespace string) {
	kind := "statefulsets"
	log := p.log
	log.Info("discoverRes start: ", kind)
	defer func() {
		log.Info("discoverRes exit: ", kind)
	}()

	appsV1 := p.clientset.AppsV1()
	api := appsV1.StatefulSets(namespace)
	itemList, err := api.List(p.ctx, metav1.ListOptions{})
	if err != nil {
		p.cc.ReportError("list-res", kind+"#"+namespace, err)
		return
	}
	total := len(itemList.Items)
	for idx, item := range itemList.Items {
		resourceName := kind + "#" + namespace + "/" + item.Name
		p.log.Printf("Resource %d/%d: %s", idx+1, total, resourceName)
		itemInfo, err := api.Get(p.ctx, item.Name, metav1.GetOptions{})
		p.reportResource(resourceName, itemInfo, err)
	}
}

func (p *Probe) discoverStorageClasses() {
	kind := "storageclasses"
	log := p.log
	log.Info("discoverRes start: ", kind)
	defer func() {
		log.Info("discoverRes exit: ", kind)
	}()

	storageV1 := p.clientset.StorageV1()
	api := storageV1.StorageClasses()
	itemList, err := api.List(p.ctx, metav1.ListOptions{})
	if err != nil {
		p.cc.ReportError("list-res", kind, err)
		return
	}
	total := len(itemList.Items)
	for idx, item := range itemList.Items {
		resourceName := kind + "#" + item.Name
		p.log.Printf("Resource %d/%d: %s", idx+1, total, resourceName)
		itemInfo, err := api.Get(p.ctx, item.Name, metav1.GetOptions{})
		p.reportResource(resourceName, itemInfo, err)
	}
}

func (p *Probe) discoverCSINodes() {
	kind := "csinodes"
	log := p.log
	log.Info("discoverRes start: ", kind)
	defer func() {
		log.Info("discoverRes exit: ", kind)
	}()

	storageV1 := p.clientset.StorageV1()
	api := storageV1.CSINodes()
	itemList, err := api.List(p.ctx, metav1.ListOptions{})
	if err != nil {
		p.cc.ReportError("list-res", kind, err)
		return
	}
	total := len(itemList.Items)
	for idx, item := range itemList.Items {
		resourceName := kind + "#" + item.Name
		p.log.Printf("Resource %d/%d: %s", idx+1, total, resourceName)
		itemInfo, err := api.Get(p.ctx, item.Name, metav1.GetOptions{})
		p.reportResource(resourceName, itemInfo, err)
	}
}

func (p *Probe) discoverCSIDrivers() {
	kind := "csidrivers"
	log := p.log
	log.Info("discoverRes start: ", kind)
	defer func() {
		log.Info("discoverRes exit: ", kind)
	}()

	storageV1 := p.clientset.StorageV1()
	api := storageV1.CSIDrivers()
	itemList, err := api.List(p.ctx, metav1.ListOptions{})
	if err != nil {
		p.cc.ReportError("list-res", kind, err)
		return
	}
	total := len(itemList.Items)
	for idx, item := range itemList.Items {
		resourceName := kind + "#" + item.Name
		p.log.Printf("Resource %d/%d: %s", idx+1, total, resourceName)
		itemInfo, err := api.Get(p.ctx, item.Name, metav1.GetOptions{})
		p.reportResource(resourceName, itemInfo, err)
	}
}

func (p *Probe) discoverCSIStorageCapacities(namespace string) {
	kind := "csistoragecapacities"
	log := p.log
	log.Info("discoverRes start: ", kind)
	defer func() {
		log.Info("discoverRes exit: ", kind)
	}()

	storageV1 := p.clientset.StorageV1()
	api := storageV1.CSIStorageCapacities(namespace)
	itemList, err := api.List(p.ctx, metav1.ListOptions{})
	if err != nil {
		p.cc.ReportError("list-res", kind+"#"+namespace, err)
		return
	}
	total := len(itemList.Items)
	for idx, item := range itemList.Items {
		resourceName := kind + "#" + namespace + "/" + item.Name
		p.log.Printf("Resource %d/%d: %s", idx+1, total, resourceName)
		itemInfo, err := api.Get(p.ctx, item.Name, metav1.GetOptions{})
		p.reportResource(resourceName, itemInfo, err)
	}
}

func (p *Probe) discoverJobs(namespace string) {
	kind := "jobs"
	log := p.log
	log.Info("discoverRes start: ", kind)
	defer func() {
		log.Info("discoverRes exit: ", kind)
	}()

	batchV1 := p.clientset.BatchV1()
	api := batchV1.Jobs(namespace)
	itemList, err := api.List(p.ctx, metav1.ListOptions{})
	if err != nil {
		p.cc.ReportError("list-res", kind+"#"+namespace, err)
		return
	}
	total := len(itemList.Items)
	for idx, item := range itemList.Items {
		resourceName := kind + "#" + namespace + "/" + item.Name
		p.log.Printf("Resource %d/%d: %s", idx+1, total, resourceName)
		itemInfo, err := api.Get(p.ctx, item.Name, metav1.GetOptions{})
		p.reportResource(resourceName, itemInfo, err)
	}
}

func (p *Probe) discoverCronJobs(namespace string) {
	kind := "cronjobs"
	log := p.log
	log.Info("discoverRes start: ", kind)
	defer func() {
		log.Info("discoverRes exit: ", kind)
	}()

	batchV1 := p.clientset.BatchV1()
	api := batchV1.CronJobs(namespace)
	itemList, err := api.List(p.ctx, metav1.ListOptions{})
	if err != nil {
		p.cc.ReportError("list-res", kind+"#"+namespace, err)
		return
	}
	total := len(itemList.Items)
	for idx, item := range itemList.Items {
		resourceName := kind + "#" + namespace + "/" + item.Name
		p.log.Printf("Resource %d/%d: %s", idx+1, total, resourceName)
		itemInfo, err := api.Get(p.ctx, item.Name, metav1.GetOptions{})
		p.reportResource(resourceName, itemInfo, err)
	}
}

func (p *Probe) discoverHorizontalPodAutoscalers(namespace string) {
	kind := "horizontalpodautoscalers"
	log := p.log
	log.Info("discoverRes start: ", kind)
	defer func() {
		log.Info("discoverRes exit: ", kind)
	}()

	autoscalingV2 := p.clientset.AutoscalingV2()
	api := autoscalingV2.HorizontalPodAutoscalers(namespace)
	itemList, err := api.List(p.ctx, metav1.ListOptions{})
	if err != nil {
		p.cc.ReportError("list-res", kind+"#"+namespace, err)
		return
	}
	total := len(itemList.Items)
	for idx, item := range itemList.Items {
		resourceName := kind + "#" + namespace + "/" + item.Name
		p.log.Printf("Resource %d/%d: %s", idx+1, total, resourceName)
		itemInfo, err := api.Get(p.ctx, item.Name, metav1.GetOptions{})
		p.reportResource(resourceName, itemInfo, err)
	}
}

func (p *Probe) discoverLeases(namespace string) {
	kind := "leases"
	log := p.log
	log.Info("discoverRes start: ", kind)
	defer func() {
		log.Info("discoverRes exit: ", kind)
	}()

	coordinationV1 := p.clientset.CoordinationV1()
	api := coordinationV1.Leases(namespace)
	itemList, err := api.List(p.ctx, metav1.ListOptions{})
	if err != nil {
		p.cc.ReportError("list-res", kind+"#"+namespace, err)
		return
	}
	total := len(itemList.Items)
	for idx, item := range itemList.Items {
		resourceName := kind + "#" + namespace + "/" + item.Name
		p.log.Printf("Resource %d/%d: %s", idx+1, total, resourceName)
		itemInfo, err := api.Get(p.ctx, item.Name, metav1.GetOptions{})
		p.reportResource(resourceName, itemInfo, err)
	}
}
