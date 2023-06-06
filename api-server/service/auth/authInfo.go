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
	"github.com/lestrrat-go/jwx/v2/jwt"
)

type AuthInfo interface {
	OrgId() string
	Jwt() jwt.Token
}

type defaultAuthInfo struct {
	jwt jwt.Token
}

func FromJwt(jwt jwt.Token) AuthInfo {
	return defaultAuthInfo{jwt: jwt}
}

func (t defaultAuthInfo) Jwt() jwt.Token {
	return t.jwt
}

func (t defaultAuthInfo) OrgId() string {
	//v, exist := t.jwt.Get("context_name")
	// if !exist {
	// 	return ""
	// }
	// return v.(string)
	return "elastic"
}
