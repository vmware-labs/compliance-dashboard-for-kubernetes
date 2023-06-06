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
package common

import (
	"collie-api-server/util"
	"context"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
	lru "k8s.io/utils/lru"
)

var (
	stateLRU *lru.Cache = lru.New(1024)
)

func init() {
}

func GetAuthUrl(oauthConfig *oauth2.Config) string {
	state := util.RandomString(8)
	stateLRU.Add(state, nil)
	return oauthConfig.AuthCodeURL(state)
}

func HandleCallback(oauthConfig *oauth2.Config, state string, code string) (*oauth2.Token, error, int) {

	_, present := stateLRU.Get(state)
	if !present {
		return nil, fmt.Errorf("Invalid state parameter: %s", state), http.StatusBadRequest
	}
	stateLRU.Remove(state)

	token, err := oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}

	return token, nil, http.StatusOK
}
