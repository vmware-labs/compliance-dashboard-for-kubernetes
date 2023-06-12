/*
Copyright 2023-2023 VMware Inc.
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
	"github.com/spf13/viper"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
)

var oauthConfig map[string]string
var muOauth sync.Mutex

func init() {
	dumpEnv()
	Require("oauth.csp.clientId")
	Require("oauth.gitlab.clientId")
}

func Require(name string) string {
	cfg := getOAuthConfig()
	lowerName := strings.ToLower(name)
	v, ok := cfg[lowerName]
	if !ok {
		panic("Missing oauth config: " + name)
	}
	if v == "<from-secret>" {
		panic("Missing oauth config (need env): " + name)
	}
	if v == "" {
		panic("Missing oauth config (empty value): " + name)
	}
	return v
}

// Get configuration bound to environment variables.
func getOAuthConfig() map[string]string {
	if oauthConfig != nil {
		return oauthConfig
	}

	muOauth.Lock()
	defer muOauth.Unlock()
	if oauthConfig == nil {
		oauthConfig = loadOAuthConfig()
	}

	return oauthConfig
}

// func copyMap(original map[string]string) map[string]string {
// 	d := make(map[string]string)
// 	for key, value := range original {
// 		d[key] = value
// 	}
// 	return d
// }

func dumpEnv() {
	for _, e := range os.Environ() {
		log.Println("ENV", e)
	}
}

func getEnviron() map[string]string {
	m := map[string]string{}
	//prefix = prefix + "_"
	for _, e := range os.Environ() {
		if i := strings.Index(e, "="); i >= 0 {
			k := strings.ToLower(e[:i])
			//if !strings.HasPrefix(k, prefix) {
			//	continue
			//}

			k = strings.ReplaceAll(k, "_", ".")
			v := e[i+1:]

			if _, exist := m[k]; exist {
				panic("Duplicated environment variable: " + k)
			}

			m[k] = v
		}
	}
	return m
}

func loadOAuthConfig() map[string]string {

	v := viper.New()
	default_config := "config/oauth-default.yaml"
	v.SetConfigFile(default_config)

	//v.SetEnvPrefix("MY_APP") // Optional: Set environment variable prefix
	//v.AutomaticEnv()        // Enable environment variable support
	//v.SetEnvKeyReplacer(strings.NewReplacer("_", "."))
	//v.AllowEmptyEnv(false)   // Allow empty environment variables

	if err := v.ReadInConfig(); err != nil {
		panic(fmt.Errorf("Error reading config file. File: %s. error: %s\n", default_config, err))
	}

	cfg := flattenMap(v.AllSettings())

	// Unify keys. Lower case only
	for k, v := range cfg {
		lowerKey := strings.ToLower(k)
		if k != lowerKey {
			cfg[lowerKey] = v
			delete(cfg, k)
		}
	}

	//Override with env variables. The default viper.AutomaticEnv handles env with only
	//the top level keys, but we want a hierarchy in the yaml config file.
	env := getEnviron()
	for k, v := range env {
		//log.Println("ENV", k, v)
		cfg[k] = v
	}

	// apply variable substitution
	for k, v := range cfg {
		cfg[k] = replaceVariables(v, cfg)

	}

	//cfg.Csp.validate("csp")
	//cfg.Gitlab.validate("gitlab")

	for k, v := range cfg {
		if strings.Contains(k, "secret") {
			log.Printf("%s=<masked> %d", k, len(v))
			continue
		}
		log.Printf("%s=%s", k, v)
	}

	return cfg
}

func flattenMap(nestedMap map[string]interface{}) map[string]string {
	flatMap := make(map[string]string)
	for key, value := range nestedMap {
		if nested, ok := value.(map[string]interface{}); ok {
			flattened := flattenMap(nested)
			for flatKey, flatValue := range flattened {
				newKey := fmt.Sprintf("%s.%s", key, flatKey)
				flatMap[newKey] = flatValue
			}
		} else {
			flatMap[key], ok = value.(string)
			if !ok {
				panic("Value is not a string. Key=" + key)
			}
		}
	}
	return flatMap
}

func replaceVariables(input string, variables map[string]string) string {
	re := regexp.MustCompile(`\$\{(.+)\}`)
	output := re.ReplaceAllStringFunc(input, func(match string) string {
		variableName := match[2 : len(match)-1]
		if value, ok := variables[variableName]; ok {
			return value
		}
		// If the variable is not found in the map, return the original match
		return match
	})
	return output
}

// func (o *OAuthEndpoint) validate(name string) {
// 	if o.AuthURL == "" {
// 		panic(fmt.Errorf("Missing AuthURL in oauth config '%s'", name))
// 	}
// 	if o.TokenURL == "" {
// 		panic(fmt.Errorf("Missing TokenURL in oauth config '%s'", name))
// 	}
// 	if o.RedirectURL == "" {
// 		panic(fmt.Errorf("Missing RedirectURL in oauth config '%s'", name))
// 	}
// 	if o.ClientId == "" {
// 		panic(fmt.Errorf("Missing ClientId in oauth config '%s'", name))
// 	}
// 	if o.ClientSecret == "" {
// 		panic(fmt.Errorf("Missing ClientSecret in oauth config '%s'", name))
// 	}
// }
