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

package controller

import (
	"fmt"
	"log"
	"net/http"
	"text/template"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"

	"collie-api-server/config"
	"collie-api-server/httputil"
	middleware "collie-api-server/middleware"
	"collie-api-server/service/oauth/csp"
	"collie-api-server/service/oauth/gitlab"
	"collie-api-server/service/oauth/google"
	"collie-api-server/service/org"
)

var homeUrl string = config.Require("collie.url") + "/collie/portal"

func (c *Controller) PortalIndex(ctx *gin.Context) {

	authInfo := middleware.GetAuth(ctx)
	orgInfo, err := org.EnsureOnboard(authInfo.OrgId())
	if err != nil {
		httputil.Abort(ctx, http.StatusInternalServerError, err)
		return
	}
	cfg := config.Get()
	data := gin.H{
		"grafanaURL":   cfg.GrafanaURL,       //"https://collie.eng.vmware.com/d/qIbLYbT4z/k8s-compliance-report"
		"grafanaOrgId": orgInfo.GrafanaOrgId, //"1"
	}
	ctx.HTML(http.StatusOK, "index.html", data)
}

func (c *Controller) PortalLogin(ctx *gin.Context) {
	data := map[string]interface{}{
		"cspAuthUrl":    csp.GetAuthUrl(),
		"gitlabAuthUrl": gitlab.GetAuthUrl(),
		"googleAuthUrl": google.GetAuthUrl(),
	}
	//ctx.HTML(http.StatusOK, "login.html", data)
	renderTemplate(ctx, "login.html", data)
}

func renderTemplate(c *gin.Context, templateName string, data gin.H) {
	tmpl := template.Must(template.ParseFiles("assets/" + templateName))

	// Disable auto-escaping of template variables
	//tmpl.Option("html").EscapeHTML = false

	err := tmpl.ExecuteTemplate(c.Writer, templateName, data)
	if err != nil {
		// Handle the error
		c.AbortWithError(http.StatusInternalServerError, err)
	}
}

type fnCallback func(string, string) (*oauth2.Token, error, int)

func callback(c *Controller, handleCallback fnCallback, provider string, ctx *gin.Context) {
	state := ctx.Query("state")
	code := ctx.Query("code")
	token, err, statusCode := handleCallback(state, code)
	if err != nil {
		log.Printf("Auth failed: %s", err.Error())
		httputil.Abort(ctx, statusCode, err)
	} else {
		age := int(time.Until(token.Expiry).Seconds())
		token := provider + "/" + token.AccessToken
		ctx.SetCookie("auth", token, age, "", "", false, true)
		ctx.Redirect(http.StatusTemporaryRedirect, homeUrl)
	}
}

func (c *Controller) OauthCallback(ctx *gin.Context) {
	provider := ctx.Param("provider")
	if provider == "csp" {
		callback(c, csp.HandleCallback, provider, ctx)
	} else if provider == "gitlab" {
		callback(c, gitlab.HandleCallback, provider, ctx)
	} else if provider == "google" {
		callback(c, google.HandleCallback, provider, ctx)
	} else {
		httputil.Abort(ctx, http.StatusBadRequest, fmt.Errorf("Unknown provider: %s", provider))
	}
}
