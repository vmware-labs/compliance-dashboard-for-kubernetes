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

package auth

import (
	"reflect"
	"collie-api-server/util"
)

type AuthInfo interface {
	OrgId()								string
	Username()							string
	Get(name string)					string
	GetAny(name string)					interface{}
	Set(name string, value interface{})
	AsMap()								map[string]interface{}
}

type defaultAuthInfo struct {
	data map[string]interface{}
}

func FromMap(data map[string]interface{}) AuthInfo {
	return defaultAuthInfo{data: util.DeepCopyMap(data)}
}

func (t defaultAuthInfo) OrgId() string {
	return "elastic"
	//return t.Get("orgId")
}

func (t defaultAuthInfo) Username() string {
	return t.Get("username")
}

func (t defaultAuthInfo) Get(name string) string {
	v, ok := t.data[name]
	if !ok {
		return ""
	}
	if reflect.TypeOf(v) == reflect.TypeOf("") {
		return v.(string)
	}
	return ""
}

func (t defaultAuthInfo) GetAny(name string) interface{} {
	return t.data[name]
}

func (t defaultAuthInfo) Set(name string, value interface{}) {
	t.data[name] = value
}

func (t defaultAuthInfo) AsMap() map[string]interface{} {
	return util.DeepCopyMap(t.data)
}
