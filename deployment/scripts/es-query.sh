#!/bin/bash

ORG_ID=elastic
INDEX_URL=$ES_URL/collie-k8s-$ORG_ID/_search

curl -k -u $ES_KEY -XGET -H "Content-Type: application/json" -d '{
		"query": {
			"bool": {
				"must": [
					{
						"range": {
							"@timestamp": {
								"lt": "2023-06-08T22:52:13Z"
							}
						}
					},
					{
						"term": {
							"a": "demo"
						}
					}
				]
			}
		}
	}' $INDEX_URL

