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
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	lru "k8s.io/utils/lru"

	"collie-api-server/config"
	"collie-api-server/service/oauth/csp"
	"collie-api-server/service/oauth/gitlab"
	"collie-api-server/service/oauth/google"
)

var (
	tokenCache *lru.Cache
)

func init() {
	tokenCache = lru.New(1024)
}

func generateRandToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

func Authenticate(token string) (AuthInfo, error) {
	token = strings.TrimPrefix(token, "Bearer ")
	token = strings.TrimPrefix(token, "Token ")

	parts := strings.Split(token, "/")
	var provider string
	var code string
	if len(parts) == 1 {
		provider = "api"
		code = parts[0]
	} else {
		provider = parts[0]
		code = parts[1]
	}

	if provider == "api" {
		return validateApiToken(code)
	} else if provider == "csp" {
		t, err := csp.Validate(code)
		if err != nil {
			return nil, err
		}
		t["orgId"] = "csp/" + t["context_name"].(string)
		return FromMap(t), nil
	} else if provider == "gitlab" {
		t, err := gitlab.Validate(code)
		if err != nil {
			return nil, err
		}
		t["orgId"] = "gitlab/" + fmt.Sprintf("%v", t["id"])
		return FromMap(t), nil
	} else if provider == "google" {
		t, err := google.Validate(code)
		if err != nil {
			return nil, err
		}
		t["orgId"] = "google/" + fmt.Sprintf("%v", t["id"])
		return FromMap(t), nil
	} else {
		return nil, errors.New("Invalid auth provider: " + provider)
	}
}

func validateApiToken(code string) (AuthInfo, error) {
	authInfo, ok := tokenCache.Get(code)
	if !ok {
		return nil, errors.New("OTP does not exist")
	}
	return authInfo.(AuthInfo), nil
}

func AddToken(token string) {
	tokenCache.Add(token, nil)
}

func GenerateApiKey(authInfo AuthInfo) string {
	token := authInfo.OrgId() + ":" + generateRandToken(16)
	tokenCache.Add(token, authInfo)
	return token
}

func GenerateEsKey(authInfo AuthInfo) string {
	//esUser := authInfo.OrgId()
	//esPwd := "gen"
	//return esUser + ":" + esPwd
	return config.Get().EsKey
}
