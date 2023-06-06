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

package controller

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"collie-api-server/config"
	"collie-api-server/middleware"
	"collie-api-server/service"
	"collie-api-server/service/auth"
	"collie-api-server/service/es"
	"collie-api-server/util"
)

// GetBootstrap godoc
//
//	@Summary		Get command line for bootstrapping
//	@Description	Return command line used for bootstraping
//	@Tags			onboarding
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	string
//	@Failure		400	{object}	httputil.HTTPError
//	@Failure		404	{object}	httputil.HTTPError
//	@Failure		500	{object}	httputil.HTTPError
//	@Router			/onboarding/bootstrap [get]
func (c *Controller) GetBootstrap(ctx *gin.Context) {
	cfg := config.Get()
	authInfo := middleware.GetAuth(ctx)
	token := auth.GenerateApiKey(authInfo)
	agentId := util.RandomString(8)
	collieUrl := fmt.Sprintf("%s/api/v1/onboarding/agent.yaml?provider=AKS&aid=%s", cfg.ApiURL, agentId)
	cmd := fmt.Sprintf("curl -skH \"Authorization: Bearer %s\" \"%s\" | kubectl apply -f -", token, collieUrl)
	data := map[string]string{
		"aid": agentId,
		"cmd": cmd,
	}
	ctx.JSON(http.StatusOK, data)
}

// GetOnboardingStatus godoc
//
//	@Summary		Get onboarding status
//	@Description	Get onboarding status
//	@Tags			onboarding
//	@Accept			json
//	@Produce		json
//	@Param			aid	query		string	false	"Agent id"
//	@Success		200	{object}	string
//	@Failure		400	{object}	httputil.HTTPError
//	@Failure		404	{object}	httputil.HTTPError
//	@Failure		500	{object}	httputil.HTTPError
//	@Router			/onboarding/status [get]
func (c *Controller) GetOnboardingStatus(ctx *gin.Context) {
	authInfo := middleware.GetAuth(ctx)
	orgId := authInfo.OrgId()

	agentId := ctx.Query("aid")
	var hasActivity bool
	var err error
	if isTestAgentStatusSimulationEnabled() {
		hasActivity = true
	} else {
		hasActivity, err = es.HasActivities(orgId, agentId)
	}

	var status string
	if err != nil {
		status = "error: " + err.Error()
	} else if hasActivity {
		status = "connected"
	} else {
		status = "pending"
	}

	data := map[string]string{
		"orgId": orgId,
		"agent": status,
	}
	ctx.JSON(http.StatusOK, data)
}

func isTestAgentStatusSimulationEnabled() bool {
	for _, v := range os.Environ() {
		if v == "SIMULATE_AGENT_READY=1" {
			log.Printf("SIMULATE_AGENT_READY=1")
			return true
		}
	}
	return false
}

// GetAgentYaml godoc
//
//	@Summary		Get agent installation yaml file used by kubectl, for the current user.
//	@Description	Return the agent yaml file
//	@Tags			onboarding
//	@Accept			json
//	@Produce		text/plain
//	@Param			provider	query		string	false	"string enums"	Enums(AKS, EKS, Other)
//	@Param			aid			query		string	true	"Agent id"
//	@Success		200			{object}	string
//	@Failure		400			{object}	httputil.HTTPError
//	@Failure		404			{object}	httputil.HTTPError
//	@Failure		500			{object}	httputil.HTTPError
//	@Router			/onboarding/agent.yaml [get]
func (c *Controller) GetAgentYaml(ctx *gin.Context) {
	provider := ctx.DefaultQuery("provider", "Other")
	agentId := ctx.Query("aid")
	authInfo := middleware.GetAuth(ctx)
	apiKey := auth.GenerateApiKey(authInfo)
	esKey := auth.GenerateEsKey(authInfo)
	text, err := service.GenerageAgentYaml(provider, apiKey, esKey, agentId)
	if err != nil {
		err2 := ctx.AbortWithError(http.StatusInternalServerError, err)
		if err2 != nil {
			log.Printf("Error aborting context: %s", err2)
		}
		return
	}
	ctx.String(http.StatusOK, text)
}
