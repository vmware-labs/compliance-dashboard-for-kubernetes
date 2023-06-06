#!/bin/bash

curl -k -u $COLLIE_ES_USERNAME:$COLLIE_ES_PASSWORD -X DELETE "$COLLIE_ES_URL/collie-k8s-cluster-info-elastic"
curl -k -u $COLLIE_ES_USERNAME:$COLLIE_ES_PASSWORD -X DELETE "$COLLIE_ES_URL/collie-k8s-compliance-elastic"
curl -k -u $COLLIE_ES_USERNAME:$COLLIE_ES_PASSWORD -X DELETE "$COLLIE_ES_URL/collie-k8s-resource-elastic"
curl -k -u $COLLIE_ES_USERNAME:$COLLIE_ES_PASSWORD -X DELETE "$COLLIE_ES_URL/collie-k8s-activity-elastic"
curl -k -u $COLLIE_ES_USERNAME:$COLLIE_ES_PASSWORD -X DELETE "$COLLIE_ES_URL/collie-k8s-elastic"
