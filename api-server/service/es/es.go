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

package es

import (
	"context"
	"crypto/tls"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"

	"collie-api-server/config"
	"github.com/elastic/go-elasticsearch/v8"
)

type EsFacade struct {
	client      *elasticsearch.Client
	typedClient *elasticsearch.TypedClient
}

var (
	es *EsFacade
)

func init() {
	cfg := config.Get()
	_, err := createEsClient(cfg.EsURL, cfg.EsKey)
	if err != nil {
		panic(err)
	}
	log.Println("es client created")
}

func parseEsToken(esToken string) (string, string, error) {
	// Trick:
	// To support envsubst with k8s deployment yaml, the env is base64 encoded.
	// To use the same .env file to run this server locally, the base64 encoded one is not accepted here.
	// So try handle both cases.
	if !strings.Contains(esToken, ":") && strings.HasSuffix(esToken, "==") {
		//base64 encoded
		sDec, err := b64.StdEncoding.DecodeString(esToken)
		if err != nil {
			return "", "", err
		}
		esToken = string(sDec)
	}

	parts := strings.Split(esToken, ":")
	if len(parts) != 2 {
		return "", "", errors.New("Invalid esToken: " + esToken)
	}
	esUsername := parts[0]
	esPassword := parts[1]
	return esUsername, esPassword, nil
}

func createEsClient(esUrl string, esToken string) (*EsFacade, error) {

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	retryBackoff := backoff.NewExponentialBackOff()

	esUsername, esPassword, err := parseEsToken(esToken)
	if err != nil {
		log.Printf("Error parsing ES token: %s", esToken)
		return nil, err
	}

	log.Println("ES URL: ", esUrl)
	cfg := elasticsearch.Config{
		Addresses: []string{
			esUrl,
		},
		Username:  esUsername,
		Password:  esPassword,
		Transport: transport,

		// Retry on 429 TooManyRequests statuses
		RetryOnStatus: []int{502, 503, 504, 429},

		// Configure the backoff function
		//
		RetryBackoff: func(i int) time.Duration {
			if i == 1 {
				retryBackoff.Reset()
			}
			return retryBackoff.NextBackOff()
		},

		// Retry up to 5 attempts
		//
		MaxRetries: 5,
	}
	client, err := elasticsearch.NewClient(cfg)

	log.Printf("Client: %s", elasticsearch.Version)
	if err != nil {
		log.Printf("Error creating ES client: %s", err)
		return nil, err
	}

	typedClient, err := elasticsearch.NewTypedClient(cfg)
	if err != nil {
		log.Printf("Error creating ES typed client: %s", err)
		return nil, err
	}
	es = &EsFacade{client, typedClient}
	return es, nil
}

type SearchTotal struct {
	Value int `json:"value"`
}
type SearchHits struct {
	Hits  []map[string]interface{} `json:"hits"`
	Total SearchTotal              `json:"total"`
}

type SearchResult struct {
	Hits SearchHits `json:"hits"`
}

func (es *EsFacade) getDoc(indexName string, filter map[string]string, size int) ([]map[string]interface{}, error) {
	// Define the search query
	list := []string{}
	for k, v := range filter {
		list = append(list, fmt.Sprintf(`{"term":{"%s":"%s"}}`, k, v))
	}
	query := fmt.Sprintf(`{
		"query": {
			"bool": { 
				"filter": [%s]
			}
		},
		"size": %d
	}`, strings.Join(list, ","), size)

	/*
		// Create a search request
		req := esapi.SearchRequest{
			Index: []string{indexName},
			Body:  esutil.NewJSONReader(query),
		}

		// Perform the search request using the Elasticsearch v8 typed client
		res, err := req.Do(context.Background(), es.client)
		if err != nil {
			log.Fatalf("Error performing search request: %s", err)
		}
		defer res.Body.Close()

		// Check for errors in the search response
		if res.IsError() {
			log.Fatalf("Error response from Elasticsearch: %s", res.Status())
		}
	*/
	res, err := es.client.Search(
		es.client.Search.WithContext(context.Background()),
		es.client.Search.WithIndex(indexName),
		es.client.Search.WithBody(strings.NewReader(query)),
	)
	if err != nil {
		log.Printf("Error parsing search response: %s", err)
		return nil, err
	}
	defer res.Body.Close()
	if res.IsError() {
		log.Printf("Error response from Elasticsearch: %s", res.Status())
		return nil, errors.New(res.Status())
	}

	// Parse the search response using the SearchResult struct
	var searchResult SearchResult
	err = json.NewDecoder(res.Body).Decode(&searchResult)
	if err != nil {
		log.Printf("Error parsing search response: %s", err)
		return nil, err
	}

	return searchResult.Hits.Hits, nil
}

func GetDoc(indexName string, filter map[string]string, size int) ([]map[string]interface{}, error) {
	return es.getDoc(indexName, filter, size)
}

func GetDoc1(indexName string, filter map[string]string) (map[string]interface{}, error) {
	docs, err := es.getDoc(indexName, filter, 1)
	if err != nil {
		return nil, err
	}
	if len(docs) == 0 {
		return nil, nil
	}
	return docs[0], nil
}

func HasActivities(orgId string, agentId string) (bool, error) {
	indexName := "collie-k8s-" + orgId
	filter := map[string]string{
		"a": agentId,
	}
	doc, err := GetDoc1(indexName, filter)
	if err != nil {
		return false, err
	}
	return doc != nil, nil
}
