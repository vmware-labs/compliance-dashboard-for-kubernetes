#!/bin/bash

for i in ./grafana-data-sources/*; do \
	curl -k -X "POST" "$COLLIE_GRAFANA_URL/api/datasources" \
    -H "Content-Type: application/json" \
     --user $COLLIE_GRAFANA_CREDENTIAL \
     --data-binary @$i
done
