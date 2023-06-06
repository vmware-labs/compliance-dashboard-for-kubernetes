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

package service

import (
	"bytes"
	"collie-api-server/config"
	_ "embed"
	b64 "encoding/base64"
	"text/template"
)

type AgentYamlParams struct {
	ApiKey   string
	ApiUrl   string
	EsUrl    string
	EsKey    string
	Provider string
	Image    string
	AgentId  string
}

var (
	//go:embed template/agent.yaml
	templateAgentYaml string

	//go:embed template/kube-bench-job.yaml
	templateKubeBenchJob string

	//go:embed template/kube-hunter-job.yaml
	templateKubeHunterJob string
)

func GenerageAgentYaml(provider string, apiKey string, esKey string, agentId string) (string, error) {

	cfg := config.Get()
	t, err := template.New("agent.yaml").
		Option("missingkey=error").
		Parse(templateAgentYaml)
	if err != nil {
		return "", err
	}

	data := AgentYamlParams{
		ApiUrl:   cfg.ApiURL,
		ApiKey:   b64encode(apiKey), // secret, need encoding
		EsUrl:    cfg.EsURL,
		EsKey:    b64encode(esKey), // secret, need encoding
		Provider: provider,
		Image:    cfg.AgentImage,
		AgentId:  agentId,
	}

	var buffer bytes.Buffer
	err = t.Execute(&buffer, data)
	if err != nil {
		return "", err
	}
	text := buffer.String()

	text = combineK8sYaml(text, templateKubeBenchJob)
	text = combineK8sYaml(text, templateKubeHunterJob)

	return text, nil
}

func combineK8sYaml(src1 string, src2 string) string {
	return src1 + "\n---\n" + src2
}

func b64encode(val string) string {
	return b64.StdEncoding.EncodeToString([]byte(val))
}
