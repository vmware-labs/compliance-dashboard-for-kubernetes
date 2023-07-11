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

package google

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"collie-api-server/config"
	"collie-api-server/service/oauth/common"
)

var (
	oauthConfig *oauth2.Config
)

func init() {

	oauthConfig = &oauth2.Config{
		ClientID:     config.Require("oauth.google.clientId"),
		ClientSecret: config.Require("oauth.google.clientSecret"),
		Endpoint: google.Endpoint,
		// oauth2.Endpoint{
		// 	AuthURL:  config.Require("oauth.google.authUrl"),
		// 	TokenURL: config.Require("oauth.google.tokenUrl"),
		// },
		RedirectURL: config.Require("oauth.google.redirectUrl"),
		Scopes: []string{
			"profile",
		},
	}
}

func GetAuthUrl() string {
	return common.GetAuthUrl(oauthConfig)
}

func Validate(accessToken string) (map[string]interface{}, error) {
	return getUserInfo(accessToken)
}

func HandleCallback(state string, code string) (*oauth2.Token, error, int) {
	token, err, status := common.HandleCallback(oauthConfig, state, code)
	if err != nil {
		return token, err, status
	}

	// Use the token to make requests to the IDP API or extract user information
	// For example, you can use the token to retrieve the user's username
	_, err = getUserInfo(token.AccessToken)
	if err != nil {
		return token, err, http.StatusInternalServerError
	}

	//fmt.Printf("Authenticated user: %v", util.ToJson(userInfo))

	return token, nil, http.StatusOK
}

func getUserInfo(accessToken string) (map[string]interface{}, error) {

	userInfoUrl := "https://www.googleapis.com/oauth2/v3/userinfo"
	headers := map[string]string {
		"Authorization": fmt.Sprintf("Bearer %s", accessToken),
	}
	userInfo, err := httpGetJson(userInfoUrl, headers)
	if err != nil {
		return nil, err
	}

	return userInfo, nil
}

func httpGetJson(targetURL string, headers map[string]string) (map[string]interface{}, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Println("Dump body", string(body))
		return nil, err
	}
	return data, nil
}

func trimUrl(fullURL string) string {
	parsedURL, err := url.Parse(fullURL)
	if err != nil {
		return fullURL
	}

	parsedURL.Path = ""     // Remove the context path
	parsedURL.RawQuery = "" // Optional: Remove query parameters if necessary
	parsedURL.Fragment = "" // Optional: Remove fragment if necessary

	return parsedURL.String()
}