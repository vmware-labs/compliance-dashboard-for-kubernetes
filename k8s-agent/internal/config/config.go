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

package config

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	Log        Log    `mapstructure:"log"`
	API        API    `mapstructure:"api"`
	ES         API    `mapstructure:"es"`
	Kubeconfig string `mapstructure:"kubeconfig"`
	AgentId    string `mapstructure:"agentId"`

	Provider string `mapstructure:"provider"`
	EKS      *EKS   `mapstructure:"eks"`
	GKE      *GKE   `mapstructure:"gke"`
	AKS      *AKS   `mapstructure:"aks"`

	Static      *Static     `mapstructure:"static"`
	Controller  *Controller `mapstructure:"controller"`
	PprofPort   int         `mapstructure:"pprof.port"`
	HealthzPort int         `mapstructure:"healthz_port"`
}

type Log struct {
	Level int `mapstructure:"level"`
}

type API struct {
	Key string `mapstructure:"key"`
	URL string `mapstructure:"url"`
}

type EKS struct {
	AccountID   string `mapstructure:"account_id"`
	Region      string `mapstructure:"region"`
	ClusterName string `mapstructure:"cluster_name"`
}

type GKE struct {
	Region      string `mapstructure:"region"`
	ProjectID   string `mapstructure:"project_id"`
	ClusterName string `mapstructure:"cluster_name"`
	Location    string `mapstructure:"location"`
}

type AKS struct {
	NodeResourceGroup string `mapstructure:"node_resource_group"`
	Location          string `mapstructure:"location"`
	SubscriptionID    string `mapstructure:"subscription_id"`
}

type Static struct {
	SkipClusterRegistration bool   `mapstructure:"skip_cluster_registration"`
	ClusterID               string `mapstructure:"cluster_id"`
	OrganizationID          string `mapstructure:"organization_id"`
}

type Controller struct {
	Interval                       time.Duration `mapstructure:"interval"`
	PrepTimeout                    time.Duration `mapstructure:"prep_timeout"`
	InitialSleepDuration           time.Duration `mapstructure:"initial_sleep_duration"`
	HealthySnapshotIntervalLimit   time.Duration `mapstructure:"healthy_snapshot_interval_limit"`
	InitializationTimeoutExtension time.Duration `mapstructure:"initialization_timeout_extension"`
}

var cfg *Config
var mu sync.Mutex

// Get configuration bound to environment variables.
func Get() Config {
	if cfg != nil {
		return *cfg
	}

	mu.Lock()
	defer mu.Unlock()
	if cfg != nil {
		return *cfg
	}

	viper.SetDefault("controller.interval", 15*time.Second)
	viper.SetDefault("controller.prep_timeout", 10*time.Minute)
	viper.SetDefault("controller.initial_sleep_duration", 30*time.Second)
	viper.SetDefault("controller.healthy_snapshot_interval_limit", 12*time.Minute)
	viper.SetDefault("controller.initialization_timeout_extension", 5*time.Minute)

	viper.SetDefault("healthz_port", 9876)

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AllowEmptyEnv(true)

	cfg = &Config{}
	bindEnvs(*cfg)
	if err := viper.Unmarshal(&cfg); err != nil {
		panic(fmt.Errorf("parsing configuration: %v", err))
	}

	if cfg.Log.Level == 0 {
		cfg.Log.Level = int(logrus.InfoLevel)
	}

	required(cfg.API.URL, "API_URL")
	required(cfg.API.Key, "API_KEY")
	required(cfg.ES.URL, "ES_URL")
	required(cfg.ES.Key, "ES_KEY")
	required(cfg.Provider, "PROVIDER")
	required(cfg.AgentId, "AGENTID")

	if !strings.HasPrefix(cfg.API.URL, "https://") && !strings.HasPrefix(cfg.API.URL, "http://") {
		cfg.API.URL = fmt.Sprintf("https://%s", cfg.API.URL)
	}

	if cfg.EKS != nil {
		if cfg.EKS.AccountID == "" {
			requiredWhenDiscoveryDisabled("EKS_ACCOUNT_ID")
		}
		if cfg.EKS.Region == "" {
			requiredWhenDiscoveryDisabled("EKS_REGION")
		}
		if cfg.EKS.ClusterName == "" {
			requiredWhenDiscoveryDisabled("EKS_CLUSTER_NAME")
		}
	}

	if cfg.AKS != nil {
		if cfg.AKS.SubscriptionID == "" {
			requiredWhenDiscoveryDisabled("AKS_SUBSCRIPTION_ID")
		}
		if cfg.AKS.Location == "" {
			requiredWhenDiscoveryDisabled("AKS_LOCATION")
		}
		if cfg.AKS.NodeResourceGroup == "" {
			requiredWhenDiscoveryDisabled("AKS_NODE_RESOURCE_GROUP")
		}
	}

	return *cfg
}

// Reset is used only for unit testing to reset configuration and rebind variables.
func Reset() {
	cfg = nil
}

func required(value string, name string) {
	if value == "" {
		panic(fmt.Errorf("env variable %s is required", name))
	}
}

func requiredWhenDiscoveryDisabled(variable string) {
	panic(fmt.Errorf("env variable %s is required when discovery is disabled", variable))
}

func bindEnvs(iface interface{}, parts ...string) {

	ifType := reflect.TypeOf(iface)
	ifValue := reflect.ValueOf(iface)
	for i := 0; i < ifType.NumField(); i++ {
		t := ifType.Field(i)
		v := ifValue.Field(i)

		tv, ok := t.Tag.Lookup("mapstructure")
		if !ok {
			continue
		}

		if v.Kind() == reflect.Ptr && v.Type().Elem().Kind() == reflect.Struct {
			v = reflect.New(v.Type().Elem()).Elem()
		}

		switch v.Kind() {
		case reflect.Struct:
			bindEnvs(v.Interface(), append(parts, tv)...)
		default:
			_ = viper.BindEnv(strings.Join(append(parts, tv), "."))
		}
	}
}
