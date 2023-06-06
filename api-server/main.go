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

package main

import (
	"context"
	"time"

	"collie-api-server/commonms"
	"collie-api-server/config"
	"collie-api-server/controller"
	_ "collie-api-server/docs"
	auth "collie-api-server/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var (
	GitCommit = "undefined"
	GitRef    = "no-ref"
)

//	@title			Collie API Server
//	@version		1.0
//	@description	This is the API server for Collie K8S compliance tool.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@host		localhost:8080
//	@BasePath	/api/v1

//	@securityDefinitions.basic	BasicAuth

//	@securityDefinitions.apikey	ApiKeyAuth
//	@in							header
//	@name						Authorization
//	@description				Description for what is this security definition being used

//	@securitydefinitions.oauth2.application	OAuth2Application
//	@tokenUrl								https://example.com/oauth/token
//	@scope.write							Grants write access
//	@scope.admin							Grants read and write access to administrative information

//	@securitydefinitions.oauth2.implicit	OAuth2Implicit
//	@authorizationUrl						https://example.com/oauth/authorize
//	@scope.write							Grants write access
//	@scope.admin							Grants read and write access to administrative information

//	@securitydefinitions.oauth2.password	OAuth2Password
//	@tokenUrl								https://example.com/oauth/token
//	@scope.read								Grants read access
//	@scope.write							Grants write access
//	@scope.admin							Grants read and write access to administrative information

//	@securitydefinitions.oauth2.accessCode	OAuth2AccessCode
//	@tokenUrl								https://example.com/oauth/token
//	@authorizationUrl						https://example.com/oauth/authorize
//	@scope.admin							Grants read and write access to administrative information

func main() {
	commonms.RunApp(run)
}

func run(cfg config.Config, log *logrus.Entry, ctx context.Context, exitCh chan error) error {
	return startRestController()
}

func startRestController() error {
	r := gin.Default()

	// allow all origins
	//r.Use(cors.Default())

	// CORS for https://foo.com and https://github.com origins, allowing:
	// - PUT and PATCH methods
	// - Origin header
	// - Credentials share
	// - Preflight requests cached for 12 hours
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://collie.eng.vmware.com", "http://localhost:8081", "*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH"},
		AllowHeaders:     []string{"Origin", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		// AllowOriginFunc: func(origin string) bool {
		//     return origin == "https://collie.eng.vmware.com.com"
		// },
		MaxAge: 12 * time.Hour,
	}))
	r.LoadHTMLGlob("assets/*.html")

	c := controller.NewController()

	root := r.Group("/collie")
	{
		root.Static("/assets", "./assets")

		apiV1 := root.Group("/api/v1")
		{
			onboarding := apiV1.Group("/onboarding")
			{
				onboarding.Use(auth.Authenticate)
				onboarding.GET("/bootstrap", c.GetBootstrap)
				onboarding.GET("/agent.yaml", c.GetAgentYaml)
				onboarding.GET("/status", c.GetOnboardingStatus)
			}
			agent := apiV1.Group("/agent")
			{
				agent.POST("/sync-start", c.SyncStart)
				agent.POST("/sync-complete", c.SyncComplete)
			}
		}

		oauth := root.Group("/oauth")
		{
			oauth.GET("/callback/:provider", c.OauthCallback)
		}

		portal := root.Group("/portal")
		{
			portal.GET("", auth.RedirectToLoginOnAuthFailure, c.PortalIndex)
			portal.GET("/login", c.PortalLogin)
		}
	}

	// accounts := v1.Group("/accounts")
	// {
	// 	accounts.GET(":id", c.ShowAccount)
	// 	accounts.GET("", c.ListAccounts)
	// 	accounts.POST("", c.AddAccount)
	// 	accounts.DELETE(":id", c.DeleteAccount)
	// 	accounts.PATCH(":id", c.UpdateAccount)
	// 	accounts.POST(":id/images", c.UploadAccountImage)
	// }
	// bottles := v1.Group("/bottles")
	// {
	// 	bottles.GET(":id", c.ShowBottle)
	// 	bottles.GET("", c.ListBottles)
	// }
	// admin := v1.Group("/admin")
	// {
	// 	admin.Use(auth())
	// 	admin.POST("/auth", c.Auth)
	// }
	// examples := v1.Group("/examples")
	// {
	// 	examples.GET("ping", c.PingExample)
	// 	examples.GET("calc", c.CalcExample)
	// 	examples.GET("groups/:group_id/accounts/:account_id", c.PathParamsExample)
	// 	examples.GET("header", c.HeaderExample)
	// 	examples.GET("securities", c.SecuritiesExample)
	// 	examples.GET("attribute", c.AttributeExample)
	// }

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return r.Run(":8080")
}
