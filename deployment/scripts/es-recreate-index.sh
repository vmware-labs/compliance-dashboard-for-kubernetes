#!/bin/bash

ORG_ID=elastic
INDEX_URL=$ES_URL/collie-k8s-$ORG_ID

curl -k -u $ES_USERNAME:$ES_PASSWORD -X DELETE "$INDEX_URL"

curl -k -u $ES_USERNAME:$ES_PASSWORD -X PUT -H "Content-Type: application/json" -d '
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

