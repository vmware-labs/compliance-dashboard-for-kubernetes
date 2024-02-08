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

package model

type ComplianceRecord struct {
	Timestamp string            `json:"@timestamp"`
	OrgId     string            `json:"orgId"`
	ClusterId string            `json:"clusterId"`
	RuleId    string            `json:"ruleId"`
	Severity  string            `json:"severity"`
	Url       string            `json:"url"`
	Data      map[string]string `json:"data"`
}

type ClusterInfo struct {
	Provider string `json:"provider"`

	Data interface{} `json:"data"`
}

type Compliance struct {
	Plugin      string `json:"plugin"`
	RuleId      string `json:"ruleId"`
	Category    string `json:"category"`
	Subcategory string `json:"subcategory"`
	Description string `json:"description"`
	Status      string `json:"status"` //FAIL, PASS, WARN
	Remediation string `json:"remediation"`
}
