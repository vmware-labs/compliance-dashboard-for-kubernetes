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

package gitlab

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"golang.org/x/oauth2"

	"collie-api-server/config"
	"collie-api-server/service/oauth/common"
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
	endpoint := trimUrl(oauthConfig.Endpoint.AuthURL)

	// tokenInfoUrl := endpoint + "/oauth/token/info?access_token=" + accessToken
	// tokenInfo, err := httpGetJson(tokenInfoUrl)
	// if err != nil {
	// 	return nil, err
	// }
	// fmt.Println(tokenInfo)

	//Valid - user token:
	/*
		{
			"resource_owner_id":2117,
			"scope":["read_user"],
			"expires_in":7200,
			"application":{
				"uid":"0a47ccd139b6de369ab1fd06723e27be904867f28df8451103db9224b123868a"
			},
			"created_at":1686086549,
			"scopes":["read_user"],
			"expires_in_seconds":7200
		}
	*/

	//Valid - OAuth app:
	/*
		{
			"resource_owner_id":null,
			"scope":["read_user"],
			"expires_in":7200,
			"application":{
				"uid":"0a47ccd139b6de369ab1fd06723e27be904867f28df8451103db9224b123868a"
			},
			"created_at":1686086222,
			"scopes":["read_user"],
			"expires_in_seconds":7200
		}
	*/
	//Invalid
	//{"error":"invalid_token","error_description":"The access token is invalid","state":"unauthorized"}

	// oauthUserInfoUrl := endpoint + "/oauth/userinfo?access_token=" + accessToken
	// oauthUserInfo, err := httpGetJson(oauthUserInfoUrl)
	// if err != nil {
	// 	return nil, err
	// }
	// fmt.Println(oauthUserInfo)

	userInfoUrl := endpoint + "/api/v4/user?access_token=" + accessToken
	userInfo, err := httpGetJson(userInfoUrl)
	if err != nil {
		return nil, err
	}
	/*
		{
			"id":2117,
			"username":"nanw",
			"name":"Nan Wang",
			"state":"active",
			"avatar_url":"https://gitlab.eng.vmware.com/uploads/-/system/user/avatar/2117/avatar.png",
			"web_url":"https://gitlab.eng.vmware.com/nanw",
			"created_at":"2017-11-08T22:34:05.821Z",
			"bio":"",
			"location":"",
			"public_email":"",
			"skype":"",
			"linkedin":"",
			"twitter":"",
			"website_url":"",
			"organization":"",
			"job_title":"",
			"pronouns":null,
			"bot":false,
			"work_information":null,
			"followers":1,
			"following":1,
			"is_followed":false,
			"local_time":null,
			"last_sign_in_at":"2022-02-05T16:18:42.255Z",
			"confirmed_at":"2017-11-08T22:34:05.807Z",
			"last_activity_on":"2023-06-06",
			"email":"nanw@vmware.com",
			"theme_id":1,
			"color_scheme_id":1,
			"projects_limit":400,
			"current_sign_in_at":"2023-01-02T18:37:26.047Z",
			"identities":[
				{"provider":"saml","extern_uid":"nanw@vmware.com","saml_provider_id":null},
				{"provider":"ldapmain","extern_uid":"cn=nan wang 71200,ou=glo_users,ou=global,ou=sites,ou=engineering,dc=vmware,dc=com","saml_provider_id":null}
			],
			"can_create_group":true,
			"can_create_project":true,
			"two_factor_enabled":true,
			"external":false,
			"private_profile":false,
			"commit_email":"nanw@vmware.com",
			"shared_runners_minutes_limit":null,
			"extra_shared_runners_minutes_limit":null
		}
	*/

	return userInfo, nil
}

func httpGetJson(targetURL string) (map[string]interface{}, error) {
	response, err := http.Get(targetURL)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
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
