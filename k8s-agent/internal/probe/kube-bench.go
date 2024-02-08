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
	"bytes"
	"context"
	_ "embed"
	"errors"
	"io"
	"regexp"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"collie-agent/internal/model"
)

func (p *Probe) DiscoverCompliance() error {
	log := p.log

	log.Info("DiscoverCompliance start")
	startTime := time.Now()
	defer func() {
		p.cc.DeleteOldDoc(startTime, "compliance")
		log.Info("DiscoverCompliance exit")
	}()

	namespace := "collie-agent"

	podName, err := findPod(p.ctx, p.clientset, namespace, "kube-bench-")
	if err != nil {
		return err
	}

	logContent, err := getPodLogs(p.ctx, p.clientset, namespace, podName)
	if err != nil {
		return err
	}

	results := parseKubeBenchLogs(logContent)

	for _, result := range results {
		p.cc.ReportCompliance(result)
	}

	return nil
}

func parseKubeBenchLogs(logContent string) []*model.Compliance {
	//[INFO] 1 Control Plane Security Configuration
	patternCategory, _ := regexp.Compile(`^\[INFO\] (\d .+)$`)
	//[INFO] 1.1 Control Plane Node Configuration Files
	patternSubcategory, _ := regexp.Compile(`^\[INFO\] (\d\.\d .+)$`)
	//[PASS] 1.1.1 Ensure that the API server pod specification file permissions are set to 644 or more restrictive (Automated)
	patternRule, _ := regexp.Compile(`^\[([A-Z]+)\] (\d\.\d\.\d+) (.+)$`)
	//== Remediations master ==
	patternRemediationSectionStart, _ := regexp.Compile("^== Remediations .+ ==$")
	//1.1.10 Run the below command (based on the file location on your system) on the control plane node.
	//For example,
	//chown root:root <path/to/cni/files>
	patternRemediationStart, _ := regexp.Compile(`^(\d\.\d\.\d) (.+)$`)
	//== Summary master ==
	patternSummaryStart, _ := regexp.Compile("^== Summary .+ ==$")
	patternUnknownSection, _ := regexp.Compile("^== .+ ==$")

	lines := strings.Split(logContent, "\n")

	currentCategory := ""
	currentSubcategory := ""
	isRemediationSecion := false
	currentResult := &model.Compliance{}

	mapId2Rule := map[string]*model.Compliance{}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		if parts := patternCategory.FindStringSubmatch(line); len(parts) > 0 {
			// start category
			currentCategory = parts[1]
			isRemediationSecion = false
			continue
		}
		if parts := patternSubcategory.FindStringSubmatch(line); len(parts) > 0 {
			// start subcategory
			currentSubcategory = parts[1]
			isRemediationSecion = false
			continue
		}
		if parts := patternRule.FindStringSubmatch(line); len(parts) > 0 {
			// rule
			ruleId := parts[2]
			mapId2Rule[ruleId] = &model.Compliance{
				Plugin:      "kube-bench",
				RuleId:      ruleId,
				Category:    currentCategory,
				Subcategory: currentSubcategory,
				Description: parts[3],
				Status:      parts[1],
				Remediation: "",
			}
			isRemediationSecion = false
			continue
		}
		if parts := patternRemediationSectionStart.FindStringSubmatch(line); len(parts) > 0 {
			isRemediationSecion = true
			continue
		}
		if parts := patternRemediationStart.FindStringSubmatch(line); len(parts) > 0 {
			ruleId := parts[1]
			desc := parts[2]
			currentResult = mapId2Rule[ruleId]
			currentResult.Remediation = desc
			continue
		}
		if parts := patternSummaryStart.FindStringSubmatch(line); len(parts) > 0 {
			isRemediationSecion = false
			continue
		}
		if parts := patternUnknownSection.FindStringSubmatch(line); len(parts) > 0 {
			isRemediationSecion = false
			continue
		}
		//Unknown line
		if isRemediationSecion {
			currentResult.Remediation += "\n" + line
		} //else discard
	}

	values := make([]*model.Compliance, 0, len(mapId2Rule))

	for _, v := range mapId2Rule {
		values = append(values, v)
	}
	return values
}

// func getJobLogs(clientset *kubernetes.Clientset, namespace string, jobName string) (string, error) {

// 	req := clientset.CoreV1().Pods(namespace).GetLogs(jobName, &v1.PodLogOptions{})

// 	podLogs, err := req.Stream(context.Background())
// 	if err != nil {
// 		return "", err
// 	}
// 	defer podLogs.Close()

// 	buf := new(bytes.Buffer)
// 	if _, err := io.Copy(buf, podLogs); err != nil {
// 		return "", err
// 	}
// 	return buf.String(), nil
// }

func findPod(ctx context.Context, clientset *kubernetes.Clientset, namespace string, prefix string) (string, error) {
	podsApi := clientset.CoreV1().Pods(namespace)
	podList, err := podsApi.List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", err
	}

	for _, pod := range podList.Items {
		if strings.HasPrefix(pod.Name, prefix) {
			return pod.Name, nil
		}
	}
	msg := "Pod not found: " + namespace + "/" + prefix
	return "", errors.New(msg)
}

func getPodLogs(ctx context.Context, clientset *kubernetes.Clientset, namespace string, podName string) (string, error) {

	req := clientset.CoreV1().Pods(namespace).GetLogs(podName, &v1.PodLogOptions{})

	podLogs, err := req.Stream(ctx)
	if err != nil {
		return "", err
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, podLogs); err != nil {
		return "", err
	}
	return buf.String(), nil
}
