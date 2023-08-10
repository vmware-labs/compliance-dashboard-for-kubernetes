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

package reporter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

func toESJson(agentId string, clusterId string, docType string, v interface{}) ([]byte, error) {
	// Marshal the value to JSON
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	// Convert the JSON keys from dots to underscores
	var jsonObj interface{}
	err = json.Unmarshal(jsonBytes, &jsonObj)
	if err != nil {
		return nil, err
	}
	convertKeysReplaceDotsWithUnderscores(jsonObj)

	ret := map[string]interface{}{
		"@timestamp": time.Now().Format(time.RFC3339),
		"a":          agentId,
		"c":          clusterId,
		docType:      jsonObj,
	}

	// Marshal the modified JSON object to bytes
	return json.MarshalIndent(ret, "", "  ")
}

func convertKeysReplaceDotsWithUnderscores(v interface{}) {
	switch v := v.(type) {
	case map[string]interface{}:
		for k, val := range v {
			convertKeysReplaceDotsWithUnderscores(val)
			newKey := strings.ReplaceAll(k, ".", "_")
			if newKey != k {
				delete(v, k)
				v[newKey] = val
			}
		}
	case []interface{}:
		for _, val := range v {
			convertKeysReplaceDotsWithUnderscores(val)
		}
	}
}

// func printHttpResponseJson(log *logrus.Entry, res *http.Response) {

// 	log.Println("HTTP response status", res.Status)
// 	body, err := io.ReadAll(res.Body)
// 	if err != nil {
// 		log.Printf("Error reading body: %v", err)
// 		return
// 	}

// 	var prettyJSON bytes.Buffer
// 	err3 := json.Indent(&prettyJSON, body, "", "\t")
// 	if err3 != nil {
// 		log.Println("JSON parse error: ", err3)
// 		return
// 	}

// 	log.Println(prettyJSON.String())
// }

// func dumpToFile(name string, data []byte) {
// 	err := os.WriteFile("/Users/nanw/collie/temp/"+name, data, 0644)
// 	if err != nil {
// 		panic(err)
// 	}
// }

func (cc CollieClient) reportImpl(indexPrefix string, docType string, resName string, data interface{}) {

	log := cc.Log

	indexName := indexPrefix + cc.orgId

	buf, err := toESJson(cc.agentId, cc.clusterId, docType, data)
	if err != nil {
		log.Infof("reportImpl: Error encoding JSON.  type=%s, res=%s, error=%s", docType, resName, err)
	}
	reader := bytes.NewReader(buf)

	res, err := cc.typedClient.Index(indexName).
		Raw(reader).
		Do(context.Background())

	if err != nil {
		//log.Infof("Doc: %s", string(buf))
		log.Warnf("Error adding document: type=%s, res=%s, %s", docType, resName, err)
	} else {
		//log.Infof("Doc: %s", string(buf))
		log.Infof("reportImpl: OK.  type=%s, res=%s, result=%s", docType, resName, res.Result)
	}
}

func (cc CollieClient) deleteDocumentsBeforeTimestamp(indexPrefix string, timestamp time.Time, docType string) {

	log := cc.Log
	indexName := indexPrefix + cc.orgId

	// Define the query to match documents before the given timestamp
	query := fmt.Sprintf(`{
		"query": {
			"bool": {
				"must": [
					{
						"range": {
							"@timestamp": {
								"lt": "%s"
							}
						}
					},
					{
						"term": {
							"a": "%s"
						}
					}
				],
				"filter": [
					{
						"exists": {
							"field": "%s"
						}
					}
				]
			}
		}
	}`, timestamp.Format(time.RFC3339), cc.agentId, docType)

	log.Printf("deleteDocumentsBeforeTimestamp(%s) - start...", docType)
	resp, err := cc.es.DeleteByQuery([]string{indexName}, strings.NewReader(query))

	if err != nil {
		log.Printf("deleteDocumentsBeforeTimestamp(%s) - Error: %s", docType, err.Error())
	} else if resp.IsError() {
		log.Printf("deleteDocumentsBeforeTimestamp(%s) - Error: %s", docType, resp.String())
	} else {
		log.Printf("deleteDocumentsBeforeTimestamp(%s) - success: %s", docType, resp.String())
	}
}
