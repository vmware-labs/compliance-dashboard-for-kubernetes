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

package middleware

import (
	"errors"
	"log"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"

	"collie-api-server/config"
	"collie-api-server/httputil"
	authSvc "collie-api-server/service/auth"
)

var loginUrl string = config.Require("oauth.portal") + "/collie/portal/login"

func handleAuth(c *gin.Context) bool {
	token := c.GetHeader("Authorization")
	if token == "" {
		t, err := c.Request.Cookie("auth")
		if err != nil {
			log.Printf("auth failed. token=%s, err=%s", token, err.Error())
			return false
		}
		t2, err := url.QueryUnescape(t.Value)
		if err != nil {
			log.Printf("auth failed. token=%s, err=%s", t, err.Error())
			return false
		}
		token = t2
	}
	authInfo, err := authSvc.Authenticate(token)

	if err != nil {
		log.Printf("auth failed. token=%s, err=%s", token, err.Error())
		return false
	}
	c.Set("auth", authInfo)
	return true
}

func RedirectToLoginOnAuthFailure(c *gin.Context) {
	log.Printf("RedirectToLoginOnAuthFailure")
	if handleAuth(c) {
		c.Next()
	} else {
		c.Redirect(http.StatusTemporaryRedirect, loginUrl)
	}
}

func Authenticate(c *gin.Context) {
	if handleAuth(c) {
		c.Next()
	} else {
		httputil.Abort(c, http.StatusUnauthorized, errors.New("Authorization failed"))
	}
}

func GetAuth(c *gin.Context) authSvc.AuthInfo {
	authInfo, exist := c.Get("auth")
	if !exist {
		// should never happen. Should be blocked by middleware
		panic("Missing auth info. Should be handled by middleware")
	}
	return authInfo.(authSvc.AuthInfo)
}
