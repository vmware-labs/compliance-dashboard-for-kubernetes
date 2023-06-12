#!/bin/bash

curl -k -u $ES_KEY -X POST -H "Content-Type: application/json" -d '
{
  "field1": "value1",
  "field2": "value2",
  "n1": {
  	"n11": {
		"v1": 1
	},
	"n12": 3
  }
}' $ES_URL/t1/_doc

