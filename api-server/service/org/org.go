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

package org

import (
	"collie-api-server/config"
	"collie-api-server/service/persist"
)

type OrgInfo struct {
	OrgId        string
	EsKey        string
	GrafanaOrgId string
}

var (
	orgColl persist.Store
)

func init() {
	orgColl = persist.Collection("org")
}

func Get(orgId string) (*OrgInfo, error) {
	ret, err := orgColl.Get(orgId)
	if err != nil {
		return nil, err
	}
	return ret.(*OrgInfo), nil
}

func EnsureOnboard(orgId string) (*OrgInfo, error) {
	orgInfo, err := Get(orgId)
	if err != nil {
		orgInfo = &OrgInfo{}
		orgInfo.EsKey = createEsTenant(orgId)
		orgInfo.GrafanaOrgId = createGrafanaTenant(orgId)
		orgColl.Put(orgId, orgInfo)
	}

	return orgInfo, nil
}

func createEsTenant(orgId string) string {
	return config.Get().EsKey
}

func createGrafanaTenant(orgId string) string {
	return "1"
}
