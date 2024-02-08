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
	"log"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	Log         Log         `mapstructure:"log"`
	Controller  *Controller `mapstructure:"controller"`
	PprofPort   int         `mapstructure:"pprof.port"`
	HealthzPort int         `mapstructure:"healthz_port"`

	AgentImage string `mapstructure:"agent_image"`
	CollieURL  string `mapstructure:"collie_url"`
	ApiURL     string `mapstructure:"api_url"`
	EsURL      string `mapstructure:"es_url"`
	EsKey      string `mapstructure:"es_key"`
	GrafanaURL string `mapstructure:"grafana_url"`
}

type Log struct {
	Level int `mapstructure:"level"`
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

	default_config := "config/app-default.yaml"
	viper.SetConfigFile(default_config)
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fail reading configuration: %v", err))
	}

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

	required(cfg.CollieURL, "COLLIE_URL")
	cfg.CollieURL = strings.TrimSuffix(cfg.CollieURL, "/")
	if cfg.ApiURL == "" {
		cfg.ApiURL = cfg.CollieURL + "/collie"
	}
	if cfg.EsURL == "" {
		// Parse the URL
		parsedURL, err := url.Parse(cfg.CollieURL)
		if err != nil {
			panic("Error parsing COLLIE_URL:" + err.Error())
		}
		//remove port if any, add new port 9200
		parsedURL.Host = fmt.Sprintf("%s:%d", parsedURL.Hostname(), 9200)
		parsedURL.Path = ""
		parsedURL.Scheme = "https"
		cfg.EsURL = parsedURL.String()
	}
	if cfg.EsKey == "" {
		log.Printf("ES_KEY not specified, try using ES secret")
		cfg.EsKey = Require("username") + ":" + Require("password")
	}
	required(cfg.ApiURL, "API_URL")
	required(cfg.EsURL, "ES_URL")
	required(cfg.AgentImage, "AGENT_IMAGE")
	required(cfg.GrafanaURL, "GRAFANA_URL")
	required_secret(cfg.EsKey, "ES_KEY")

	return *cfg
}

// Reset is used only for unit testing to reset configuration and rebind variables.
func Reset() {
	cfg = nil
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

func required(value string, name string) {
	if value == "" {
		panic(fmt.Errorf("env variable %s is required", name))
	}

	log.Println(name, value)
}

func required_secret(value string, name string) {
	if value == "" {
		panic(fmt.Errorf("env variable %s is required", name))
	}
}
