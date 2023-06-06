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

package csp

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"golang.org/x/oauth2"

	"collie-api-server/config"
	"collie-api-server/service/oauth/common"
	"collie-api-server/util"
)

var (
	oauthConfig *oauth2.Config
	jwkSet      jwk.Set
	jwksURL     string
)

func init() {
	jwksURL = config.Require("oauth.csp.jwksUrl")

	oauthConfig = &oauth2.Config{
		ClientID:     config.Require("oauth.csp.clientId"),
		ClientSecret: config.Require("oauth.csp.clientSecret"),
		Endpoint: oauth2.Endpoint{
			AuthURL:  config.Require("oauth.csp.authUrl"),
			TokenURL: config.Require("oauth.csp.tokenUrl"),
		},
		RedirectURL: config.Require("oauth.csp.redirectUrl"),
		Scopes: []string{
			"openid", "email", "profile",
		},
	}
}

func GetAuthUrl() string {
	return common.GetAuthUrl(oauthConfig)
}

func Validate(accessToken string) (jwt.Token, error) {
	return verifyJWT(accessToken)
}

func HandleCallback(state string, code string) (*oauth2.Token, error, int) {
	token, err, status := common.HandleCallback(oauthConfig, state, code)
	if err != nil {
		return token, err, status
	}

	// Verify the JWT token
	decoded, err := verifyJWT(token.AccessToken)
	if err != nil {
		return token, err, http.StatusInternalServerError
	}

	fmt.Printf("Authenticated user: %v", util.ToJson(decoded))

	return token, nil, http.StatusOK
}

func newJWKSet(jwkUrl string) jwk.Set {
	jwkCache := jwk.NewCache(context.Background())

	// register a minimum refresh interval for this URL.
	// when not specified, defaults to Cache-Control and similar resp headers
	err := jwkCache.Register(jwkUrl, jwk.WithMinRefreshInterval(10*time.Minute))
	if err != nil {
		panic("failed to register jwk location")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// fetch once on application startup
	_, err = jwkCache.Refresh(ctx, jwkUrl)
	if err != nil {
		panic("failed to fetch JWKS on startup")
	}
	// create the cached key set
	return jwk.NewCachedSet(jwkCache, jwkUrl)
}

func verifyJWT(base64Jwt string) (jwt.Token, error) {

	if jwkSet == nil {
		jwkSet = newJWKSet(jwksURL)
	}

	decoded, err := jwt.ParseString(base64Jwt, jwt.WithKeySet(jwkSet, jws.WithInferAlgorithmFromKey(true)))
	if err != nil {
		return nil, err
	}

	if decoded.Issuer() != config.Require("oauth.csp.issuer") {
		return nil, fmt.Errorf("Invalid issuer: " + decoded.Issuer())
	}
	azp, exist := decoded.Get("azp")
	if !exist {
		return nil, fmt.Errorf("Missing azp")
	}
	if azp != oauthConfig.ClientID {
		return nil, fmt.Errorf("azp (%s) does not match clientId", azp)
	}

	now := time.Now()

	nbf := decoded.NotBefore()
	if now.Before(nbf) {
		return nil, fmt.Errorf("nbf violation: %v", nbf)
	}

	iat := decoded.IssuedAt()
	if now.Before(iat) {
		return nil, fmt.Errorf("iat violation: %v", iat)
	}

	exp := decoded.Expiration()
	if now.After(exp) {
		return nil, fmt.Errorf("exp violation: %v", exp)
	}

	return decoded, nil
}
