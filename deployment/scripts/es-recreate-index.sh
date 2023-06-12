#!/bin/bash

ORG_ID=elastic
INDEX_URL=$ES_URL/collie-k8s-$ORG_ID

curl -k -u $ES_KEY -X DELETE "$INDEX_URL"

curl -k -u $ES_KEY -X PUT -H "Content-Type: application/json" -d '
{
  "mappings": {
    "properties": {
      "resource": {
        "dynamic": false,
	"properties": {
          "metadata": {
            "properties": {
              "name": {
                "type": "text"
	      },
	      "namespace": {
                "type": "text"
	      }
            }
          }
        }
      }
    }
  }
}' $INDEX_URL

