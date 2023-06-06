#!/bin/bash

DIR="grafana-dashboards"

# Iterate through dashboards using the current API Key
for dashboard_uid in $(curl -sS -u $COLLIE_GRAFANA_CREDENTIAL $COLLIE_GRAFANA_URL/api/search\?query\=\& | jq -r '.[] | select( .type | contains("dash-db")) | .uid'); do
    url=$(echo $COLLIE_GRAFANA_URL/api/dashboards/uid/$dashboard_uid | tr -d '\r')
    dashboard_json=$(curl -sS -u $COLLIE_GRAFANA_CREDENTIAL $url)
    dashboard_title=$(echo $dashboard_json | jq -r '.dashboard | .title' | sed -r 's/[ \/]+/_/g')
    dashboard_version=$(echo $dashboard_json | jq -r '.dashboard | .version')
    folder_title="$(echo $dashboard_json | jq -r '.meta | .folderTitle')"

    echo "Creating: ${DIR}/${folder_title}/${dashboard_title}_v${dashboard_version}.json"
    mkdir -p "${DIR}/${folder_title}"
    echo ${dashboard_json} | jq -r {meta:.meta}+.dashboard > "${DIR}/${folder_title}/${dashboard_title}_v${dashboard_version}.json"
done