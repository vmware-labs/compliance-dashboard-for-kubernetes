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
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SyncStart godoc
//
//	@Summary		Indicating a sync has been started
//	@Description	Indicating a sync has been started
//	@Tags			agent
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	string
//	@Failure		400	{object}	httputil.HTTPError
//	@Failure		404	{object}	httputil.HTTPError
//	@Failure		500	{object}	httputil.HTTPError
//	@Router			/agent/sync-start [post]
func (c *Controller) SyncStart(ctx *gin.Context) {
	log.Printf("SyncStart")
	ctx.String(http.StatusOK, "")
}

// PostDiscoveryComplete godoc
//
//	@Summary		Indicating a sync has complete
//	@Description	Indicating a sync has complete
//	@Tags			agent
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	string
//	@Failure		400	{object}	httputil.HTTPError
//	@Failure		404	{object}	httputil.HTTPError
//	@Failure		500	{object}	httputil.HTTPError
//	@Router			/agent/sync-complete [post]
func (c *Controller) SyncComplete(ctx *gin.Context) {
	log.Printf("SyncComplete")
	ctx.String(http.StatusOK, "")
}
