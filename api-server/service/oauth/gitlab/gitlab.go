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

package gitlab

import (
	"fmt"
	"net/http"

	"golang.org/x/oauth2"

	"github.com/lestrrat-go/jwx/v2/jwt"

	"collie-api-server/config"
	"collie-api-server/service/oauth/common"
	"collie-api-server/util"
)

var (
	oauthConfig *oauth2.Config
)

func init() {

	oauthConfig = &oauth2.Config{
		ClientID:     config.Require("oauth.gitlab.clientId"),
		ClientSecret: config.Require("oauth.gitlab.clientSecret"),
		Endpoint: oauth2.Endpoint{
			AuthURL:  config.Require("oauth.gitlab.authUrl"),
			TokenURL: config.Require("oauth.gitlab.tokenUrl"),
		},
		RedirectURL: config.Require("oauth.gitlab.redirectUrl"),
		Scopes: []string{
			"read_user",
		},
	}
}

func GetAuthUrl() string {
	return common.GetAuthUrl(oauthConfig)
}

func Validate(accessToken string) (jwt.Token, error) {
	return getUserInfo(accessToken)
}

func HandleCallback(state string, code string) (*oauth2.Token, error, int) {
	token, err, status := common.HandleCallback(oauthConfig, state, code)
	if err != nil {
		return token, err, status
	}

	// Use the token to make requests to the IDP API or extract user information
	// For example, you can use the token to retrieve the user's username
	userInfo, err := getUserInfo(token.AccessToken)
	if err != nil {
		return token, err, http.StatusInternalServerError
	}

	fmt.Printf("Authenticated user: %v", util.ToJson(userInfo))

	return token, nil, http.StatusOK
}

func getUserInfo(accessToken string) (jwt.Token, error) {
	// You'll need to implement the logic to make a request to the IDP API
	// and retrieve the user's username using the provided token
	// In this example, we'll just return a dummy username
	t := jwt.New()
	err := t.Set("context_name", "org1")
	return t, err
}
