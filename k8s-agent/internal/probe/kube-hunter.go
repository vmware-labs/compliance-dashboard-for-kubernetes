package probe

import (
	"bytes"
	"time"
	"context"
	"encoding/json"
	"errors"
	"io"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"strings"

	"collie-agent/internal/model"
)

type kubeHunterRecord struct {
	Location      string `json:"location"`
	Vid           string `json:"vid"`
	Category      string `json:"category"`
	Severity      string `json:"severity"`
	Vulnerability string `json:"vulnerability"`
	Description   string `json:"description"`
	Evidence      string `json:"evidence"`
	Avd_reference string `json:"avd_reference"`
	Hunter        string `json:"hunter"`
}

type kubeHunterResult struct {
	Nodes           []*kubeHunterResultNode
	Services        []*kubeHunterResultService
	Vulnerabilities []*kubeHunterRecord
}

type kubeHunterResultNode struct {
	Type     string `json:"type"`
	Location string `json:"location"`
}

type kubeHunterResultService struct {
	Service  string `json:"service"`
	Location string `json:"location"`
}

func (p *Probe) DiscoverComplianceForHunter() error {
	log := p.log

	log.Info("DiscoverCompliance start for kube-hunter")
	startTime := time.Now()
	defer func() {
		p.cc.DeleteOldDoc(startTime, "compliance")
		log.Info("DiscoverCompliance exit for kube-hunter")
	}()
	namespace := "collie-agent"

	podName, err := findPodForHunter(p.ctx, p.clientset, namespace, "kube-hunter-")
	if err != nil {
		return err
	}

	waitForKubeHunterComplete()

	logContent, err := getPodLogsForHunter(p.ctx, p.clientset, namespace, podName)

	if err != nil {
		return err
	}

	results, err := getPluginResult(logContent)
	if err != nil {
		return err
	}

	for _, result := range results {
		p.cc.ReportCompliance(result)
	}

	return nil
}

func waitForKubeHunterComplete() {
	//TODO: wait for job completion
	time.Sleep(30 * time.Second)
}

func getPluginResult(logContent string) ([]*model.Compliance, error) {
	vulnerabilities, err := parseKubeHunterLogs(logContent)
	if err != nil {
		return nil, err
	}
	var result = make([]*model.Compliance, len(vulnerabilities))
	for i := range vulnerabilities {
		categories := strings.Split(vulnerabilities[i].Category, "//")
		result[i] = &model.Compliance{
			Plugin:      "kube-hunter",
			RuleId:      vulnerabilities[i].Vid,
			Category:    categories[0],
			Subcategory: categories[1],
			Description: vulnerabilities[i].Description,
			Status:      "WARN",
			Remediation: "",
		}
	}
	return result, nil
}

func parseKubeHunterLogs(logContent string) ([]*kubeHunterRecord, error) {
	var result kubeHunterResult
	var hunterJson = lastLine(logContent)
	err := json.Unmarshal([]byte(hunterJson), &result)
	return result.Vulnerabilities, err
}

func lastLine(logContent string) string {
	var lines = strings.Split(logContent, "\n")
	var hunterJson = lines[len(lines)-2]
	return hunterJson
}

func findPodForHunter(ctx context.Context, clientset *kubernetes.Clientset, namespace string, prefix string) (string, error) {
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

func getPodLogsForHunter(ctx context.Context, clientset *kubernetes.Clientset, namespace string, podName string) (string, error) {

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
