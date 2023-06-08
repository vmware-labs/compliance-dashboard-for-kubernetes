#!/bin/bash

curl -k -u $ES_USERNAME:$ES_PASSWORD -X DELETE "$ES_URL/collie-k8s-cluster-info-elastic"
curl -k -u $ES_USERNAME:$ES_PASSWORD -X DELETE "$ES_URL/collie-k8s-compliance-elastic"
curl -k -u $ES_USERNAME:$ES_PASSWORD -X DELETE "$ES_URL/collie-k8s-resource-elastic"
curl -k -u $ES_USERNAME:$ES_PASSWORD -X DELETE "$ES_URL/collie-k8s-activity-elastic"
#curl -k -u $ES_USERNAME:$ES_PASSWORD -X DELETE "$ES_URL/collie-k8s-elastic"
