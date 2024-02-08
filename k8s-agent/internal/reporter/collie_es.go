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

package reporter

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"strings"
	"time"

	//"runtime"
	"sync/atomic"

	"github.com/cenkalti/backoff/v4"
	"github.com/dustin/go-humanize"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"

	"collie-agent/internal/model"
)

var (
	indexPrefix string = "collie-k8s-"
	numWorkers  int
	flushBytes  int
)

func init() {
	workers := 2 //runtime.NumCPU()
	flag.IntVar(&numWorkers, "workers", workers, "Number of indexer workers")
	flag.IntVar(&flushBytes, "flush", 5e+6, "Flush threshold in bytes")
	flag.Parse()
}

type CollieClient struct {
	Log         *logrus.Entry
	orgId       string
	agentId     string
	clusterId   string
	es          *elasticsearch.Client
	typedClient *elasticsearch.TypedClient
	rest        *resty.Client
}

func New(log *logrus.Entry, agentId string, clusterId string, apiUrl string, apiToken string, esUrl string, esToken string) (*CollieClient, error) {

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	retryBackoff := backoff.NewExponentialBackOff()

	parts := strings.Split(esToken, ":")
	if len(parts) != 2 {
		return nil, errors.New("Invalid esToken: " + esToken)
	}
	esUsername := parts[0]
	esPassword := parts[1]
	log.Info("ES URL: ", esUrl)
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
	es, err := elasticsearch.NewClient(cfg)

	log.Printf("Client: %s", elasticsearch.Version)
	if err != nil {
		log.Warnf("Error creating ES client: %s", err)
		return nil, err
	}

	typedClient, err := elasticsearch.NewTypedClient(cfg)
	if err != nil {
		log.Warnf("Error creating ES typed client: %s", err)
		return nil, err
	}

	log.Info("API URL: ", apiUrl)

	restClient := resty.New()
	restClient.SetTransport(transport)
	restClient.SetBaseURL(apiUrl)
	restClient.SetAuthToken(apiToken)
	orgId := esUsername
	client := CollieClient{log, orgId, agentId, clusterId, es, typedClient, restClient}
	return &client, err
}

func (cc CollieClient) Info() error {
	err := cc.apiInfo()
	if err != nil {
		return err
	}
	err = cc.esInfo()
	if err != nil {
		return err
	}
	cc.ReportActivity("cycle-start", "")
	return nil
}

func (cc CollieClient) apiInfo() error {
	log := cc.Log

	log.Info("Test connectivity to Collie API...")
	resp, err := cc.rest.R().
		Get("/api/v1/onboarding/status")
	if err != nil {
		return err
	}
	statusCode := resp.RawResponse.StatusCode
	log.Info("status code: ", statusCode)
	body := string(resp.Body())
	log.Info("response: ", body)
	if statusCode != 200 {
		return errors.New("Fail bootstrap health check: fail invoking Collie API")
	}
	return nil
}

func (cc CollieClient) esInfo() error {
	log := cc.Log
	log.Info("Test connectivity to Collie ES...")

	{
		res, err := cc.es.Info()
		if err != nil {
			return err
		}
		log.Println("info:", res.String())
	}

	{
		// Get the cluster health information
		res, err := cc.es.Cluster.Health(
			cc.es.Cluster.Health.WithContext(context.Background()),
			cc.es.Cluster.Health.WithPretty(),
		)
		if err != nil {
			return err
		}
		log.Println("cluster health:", res.String())
	}

	return nil
}

func (cc CollieClient) ReportClusterInfo(info model.ClusterInfo) {
	docType := "cluster"
	cc.reportImpl(indexPrefix, docType, "", info)
}

func (cc CollieClient) ReportResource(name string, data interface{}) {
	cc.reportImpl(indexPrefix, "resource", name, data)
}

type Activity struct {
	Operation string `json:"operation"`
	Resource  string `json:"resource"`
	Error     string `json:"error"`
}

func (cc CollieClient) ReportActivity(operation string, resource string) {
	data := Activity{
		Operation: operation,
		Resource:  resource,
	}
	cc.reportImpl(indexPrefix, "activity", "", data)
}

func (cc CollieClient) ReportError(operation string, resource string, e error) {
	cc.Log.Warnf("ReportError: operation=%s, res=%s, err=%s", operation, resource, e)
	data := Activity{
		Operation: operation,
		Resource:  resource,
		Error:     e.Error(),
	}
	cc.reportImpl(indexPrefix, "activity", "", data)
}

func (cc CollieClient) ReportCompliance(data *model.Compliance) {
	cc.reportImpl(indexPrefix, "compliance", "", data)
}

func (cc CollieClient) ReportBulk(docs []*any) {
	log := cc.Log
	es := cc.es

	var (
		countSuccessful uint64

		res *esapi.Response
		err error
	)

	indexName := indexPrefix + cc.orgId
	bi, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Index:         indexName,        // The default index name
		Client:        es,               // The Elasticsearch client
		NumWorkers:    numWorkers,       // The number of worker goroutines
		FlushBytes:    int(flushBytes),  // The flush threshold in bytes
		FlushInterval: 30 * time.Second, // The periodic flush interval
	})
	if err != nil {
		log.Warnf("Error creating the indexer: %s", err)
	}
	// <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<

	// Re-create the index
	//
	res, err = es.Indices.Create(indexName)
	if err != nil {
		log.Warnf("Cannot create index: %s", err)
	}
	if res.IsError() {
		log.Warnf("Cannot create index: %s", res)
	}
	res.Body.Close()

	start := time.Now().UTC()

	// Loop over the collection
	//
	for _, a := range docs {
		// Prepare the data payload: encode article to JSON
		//
		data, err := json.Marshal(a)
		if err != nil {
			log.Warnf("Cannot encode article %v: %s", a, err)
		}

		// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
		//
		// Add an item to the BulkIndexer
		//
		err = bi.Add(
			context.Background(),
			esutil.BulkIndexerItem{
				// Action field configures the operation to perform (index, create, delete, update)
				Action: "index",

				// DocumentID is the (optional) document ID
				//DocumentID: strconv.Itoa(a.ID),

				// Body is an `io.Reader` with the payload
				Body: bytes.NewReader(data),

				// OnSuccess is called for each successful operation
				OnSuccess: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem) {
					atomic.AddUint64(&countSuccessful, 1)
				},

				// OnFailure is called for each failed operation
				OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
					if err != nil {
						log.Printf("ERROR: %s", err)
					} else {
						log.Printf("ERROR: %s: %s", res.Error.Type, res.Error.Reason)
					}
				},
			},
		)
		if err != nil {
			log.Warnf("Unexpected error: %s", err)
		}
		// <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<
	}

	// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
	// Close the indexer
	//
	if err := bi.Close(context.Background()); err != nil {
		log.Warnf("Unexpected error: %s", err)
	}
	// <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<

	biStats := bi.Stats()

	// Report the results: number of indexed docs, number of errors, duration, indexing rate
	//
	log.Println(strings.Repeat("▔", 65))

	dur := time.Since(start)

	if biStats.NumFailed > 0 {
		log.Warnf(
			"Indexed [%s] documents with [%s] errors in %s (%s docs/sec)",
			humanize.Comma(int64(biStats.NumFlushed)),
			humanize.Comma(int64(biStats.NumFailed)),
			dur.Truncate(time.Millisecond),
			humanize.Comma(int64(1000.0/float64(dur/time.Millisecond)*float64(biStats.NumFlushed))),
		)
	} else {
		log.Printf(
			"Sucessfuly indexed [%s] documents in %s (%s docs/sec)",
			humanize.Comma(int64(biStats.NumFlushed)),
			dur.Truncate(time.Millisecond),
			humanize.Comma(int64(1000.0/float64(dur/time.Millisecond)*float64(biStats.NumFlushed))),
		)
	}
}

type HookReportSuccess struct {
}

type HookReportError struct {
}

func (cc CollieClient) ReportCompletion() {

	cc.Log.Info("ReportCompletion")
	// POST Struct, default is JSON content type. No need to set one
	resp, err := cc.rest.R().
		SetResult(&HookReportSuccess{}). // or SetResult(AuthSuccess{}).
		SetError(&HookReportError{}).    // or SetError(AuthError{}).
		Post("/api/v1/agent/sync-complete")

	// Explore response object
	fmt.Println("Response Info:")
	fmt.Println("  Error      :", err)
	fmt.Println("  Status Code:", resp.StatusCode())
	fmt.Println("  Status     :", resp.Status())
	fmt.Println("  Proto      :", resp.Proto())
	fmt.Println("  Time       :", resp.Time())
	fmt.Println("  Received At:", resp.ReceivedAt())
	fmt.Println("  Body       :\n", resp)
	fmt.Println()

	// Explore trace info
	fmt.Println("Request Trace Info:")
	ti := resp.Request.TraceInfo()
	fmt.Println("  DNSLookup     :", ti.DNSLookup)
	fmt.Println("  ConnTime      :", ti.ConnTime)
	fmt.Println("  TCPConnTime   :", ti.TCPConnTime)
	fmt.Println("  TLSHandshake  :", ti.TLSHandshake)
	fmt.Println("  ServerTime    :", ti.ServerTime)
	fmt.Println("  ResponseTime  :", ti.ResponseTime)
	fmt.Println("  TotalTime     :", ti.TotalTime)
	fmt.Println("  IsConnReused  :", ti.IsConnReused)
	fmt.Println("  IsConnWasIdle :", ti.IsConnWasIdle)
	fmt.Println("  ConnIdleTime  :", ti.ConnIdleTime)
	fmt.Println("  RequestAttempt:", ti.RequestAttempt)
	fmt.Println("  RemoteAddr    :", ti.RemoteAddr)

	cc.ReportActivity("cycle-complete", "")
}

func (cc CollieClient) DeleteOldDoc(before time.Time, docType string) {
	cc.deleteDocumentsBeforeTimestamp(indexPrefix, before, docType)
}
