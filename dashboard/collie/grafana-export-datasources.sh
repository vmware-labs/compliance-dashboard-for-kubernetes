#!/bin/bash

mkdir -p ./grafana-data-sources
curl -k "$COLLIE_GRAFANA_URL/api/datasources"  -u $COLLIE_GRAFANA_CREDENTIAL |jq -c -M '.[]'|split -l 1 - ./grafana-data-sources/
