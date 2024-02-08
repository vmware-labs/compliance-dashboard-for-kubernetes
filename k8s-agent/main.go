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

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	//"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"

	//"k8s.io/metrics/pkg/client/clientset/versioned"

	"collie-agent/internal/commonms"
	"collie-agent/internal/config"
	"collie-agent/internal/probe"
	"collie-agent/internal/reporter"
)

// type Greeter struct{}

// func (g *Greeter) Hello(ctx context.Context, req *proto.HelloRequest, rsp *proto.HelloResponse) error {
// 	rsp.Greeting = "Hello " + req.Name
// 	return nil
// }

// Set by `go build` during a release
var (
	GitCommit = "undefined"
	GitRef    = "no-ref"
)

func main() {
	commonms.RunApp(run)
}

func run(cfg config.Config, log *logrus.Entry, ctx context.Context, exitCh chan error) error {
	restconfig, err := retrieveKubeConfig(log, cfg)
	if err != nil {
		return err
	}

	if err := v1beta1.AddToScheme(scheme.Scheme); err != nil {
		return fmt.Errorf("adding metrics objs to scheme: %w", err)
	}

	restconfig.NegotiatedSerializer = serializer.NewCodecFactory(scheme.Scheme)

	clientset, err := kubernetes.NewForConfig(restconfig)
	if err != nil {
		return err
	}

	// metricsClient, err := versioned.NewForConfig(restconfig)
	// if err != nil {
	// 	return fmt.Errorf("initializing metrics client: %w", err)
	// }

	// dynamicClient, err := dynamic.NewForConfig(restconfig)
	// if err != nil {
	// 	return fmt.Errorf("initializing dynamic client: %w", err)
	// }

	//discoveryService := discovery.New(clientset, dynamicClient)

	return loop(ctx, log, cfg, clientset)

}

// func test(cc reporter.CollieClient) {
// 	cc.Info()

// 	// Test
// 	clusterInfo := model.ClusterInfo{
// 		Provider : "AKS",
// 	}
// 	cc.ReportClusterInfo(clusterInfo)
// 	complianceRecord := model.ComplianceRecord{
// 		RuleId: "rule-1",
// 	}
// 	cc.ReportCompliance(&complianceRecord)
// }

func loop(ctx context.Context, log *logrus.Entry, cfg config.Config, clientset *kubernetes.Clientset) error {

	clusterId, err := probe.GetClusterId(ctx, log, clientset)
	if err != nil {
		return fmt.Errorf("Error retrieving cluster ID: %w", err)
	}

	cc, err := reporter.New(log, cfg.AgentId, clusterId, cfg.API.URL, cfg.API.Key, cfg.ES.URL, cfg.ES.Key)
	if err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("Error creating collie client: %w", err)
	}

	p := probe.New(ctx, log, clientset, cc)
	// Test connectivity
	err = cc.Info()
	if err != nil {
		return err
	}

	// test(cc)

	for {

		err := p.DiscoverCluster()
		if err != nil {
			cc.ReportError("DiscoverCluster", "", err)
		}
		err = p.DiscoverResources()
		if err != nil {
			cc.ReportError("DiscoverResources", "", err)
		}
		err = p.DiscoverCompliance()
		if err != nil {
			cc.ReportError("DiscoverCompliance", "", err)
		}
		err = p.DiscoverComplianceForHunter()
		if err != nil {
			cc.ReportError("DiscoverComplianceForHunter", "", err)
		}

		cc.ReportCompletion()
		log.Infoln("Sleeping")
		time.Sleep(12 * time.Hour)
	}
}

func kubeConfigFromPath(kubepath string) (*rest.Config, error) {
	if kubepath == "" {
		return nil, nil
	}

	data, err := os.ReadFile(kubepath)
	if err != nil {
		return nil, fmt.Errorf("reading kubeconfig at %s: %w", kubepath, err)
	}

	restConfig, err := clientcmd.RESTConfigFromKubeConfig(data)
	if err != nil {
		return nil, fmt.Errorf("building rest config from kubeconfig at %s: %w", kubepath, err)
	}

	return restConfig, nil
}

func retrieveKubeConfig(log logrus.FieldLogger, cfg config.Config) (*rest.Config, error) {
	kubeconfig, err := kubeConfigFromPath(cfg.Kubeconfig)
	if err != nil {
		return nil, err
	}

	if kubeconfig != nil {
		log.Debug("using kubeconfig from env variables")
		return kubeconfig, nil
	}

	inClusterConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	log.Debug("using in cluster kubeconfig")
	return inClusterConfig, nil
}
