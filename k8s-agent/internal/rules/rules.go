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

package rules

import (
	"collie-agent/internal/model"
	"context"
	"log"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	K8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type fnReport func(*model.ComplianceRecord) error

func RunAll(log *logrus.Entry, clientset *kubernetes.Clientset, report fnReport) error {

	coreV1 := clientset.CoreV1()

	namespaceList, err := coreV1.Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	ignoredNamespaces := map[string]int{"kube-node-lease": 1, "kube-public": 1, "kube-system": 1}
	for _, namespace := range namespaceList.Items {
		log.Infoln("namespace", namespace.Name)
		podApi := coreV1.Pods(namespace.Name)
		podList, err := podApi.List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return err
		}
		log.Printf("There are %d pods in %s", len(podList.Items), namespace.Name)

		if _, ok := ignoredNamespaces[namespace.Name]; ok {
			log.Info("Ignore namespace ", namespace.Name)
			continue
		}

		for idx, pod := range podList.Items {
			log.Printf("pod %d %s", idx, pod.Name)

			podInfo, err := podApi.Get(context.Background(), pod.Name, metav1.GetOptions{})
			if K8sErrors.IsNotFound(err) {
				log.Printf("Pod not found: %s/%s", namespace.Name, pod.Name)
			} else if statusError, isStatus := err.(*K8sErrors.StatusError); isStatus {
				log.Printf("Error getting pod %v", statusError.ErrStatus.Message)
			} else if err != nil {
				return err
			} else {
				//log.Info(podInfo)

				runSinglePodRules(podInfo, report)
			}
		}
	}
	return nil
}

type fnPodRule func(*v1.Pod) *model.ComplianceRecord

func runSinglePodRules(pod *v1.Pod, fn fnReport) {

	podRules := []fnPodRule{
		ruleDeprecateHostPort,
		ruleDeprecateHostIp,
	}

	for _, rule := range podRules {
		ret := rule(pod)
		if ret == nil {
			continue
		}

		decoratePodRecord(pod, ret)
		err := fn(ret)
		if err != nil {
			log.Printf("Fail reporting %s", err)
		}
	}
}

func decoratePodRecord(pod *v1.Pod, r *model.ComplianceRecord) {
	if r.Data == nil {
		r.Data = make(map[string]string)
		r.Data["pod"] = pod.Name
		r.Data["namespace"] = pod.Namespace
	}
}

func ruleDeprecateHostPort(pod *v1.Pod) *model.ComplianceRecord {
	for _, c := range pod.Spec.Containers {
		for _, p := range c.Ports {
			if p.HostPort > 0 {
				return &model.ComplianceRecord{
					RuleId:   "deprecate-host-port",
					Severity: "INFO",
				}
			}
		}
	}
	return nil
}

func ruleDeprecateHostIp(pod *v1.Pod) *model.ComplianceRecord {
	for _, c := range pod.Spec.Containers {
		for _, p := range c.Ports {
			if len(p.HostIP) > 0 {
				return &model.ComplianceRecord{
					RuleId:   "deprecate-host-ip",
					Severity: "INFO",
				}
			}
		}
	}
	return nil
}
