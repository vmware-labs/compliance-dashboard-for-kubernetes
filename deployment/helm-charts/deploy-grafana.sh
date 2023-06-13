#!/bin/bash

helm list -n collie-server -a | grep grafana > /dev/null 2>&1
if [ $? -eq 0 ]; then
	helm uninstall -n collie-server grafana --wait
fi


export ES_PASSWORD=$(kubectl get -n collie-server secret/elasticsearch-master-credentials -o jsonpath="{.data.password}" | base64 -d)

envsubst < ./grafana/values-custom.yaml > ./grafana/tmp.yaml
helm install -n collie-server grafana ./grafana -f ./grafana/values.yaml -f ./grafana/tmp.yaml --wait
rm ./grafana/tmp.yaml

sleep 10

source ./kill-processes-by-keyword.sh "kubectl port-forward -n collie-server --address 0.0.0.0 services/grafana"
kubectl port-forward -n collie-server --address 0.0.0.0 services/grafana 3000:3000 &

sleep 5

GRAFANA_PWD=$(kubectl get secret --namespace collie-server grafana -o jsonpath="{.data.admin-password}" | base64 --decode)


curl -X POST \
  -H "Content-Type: application/json" \
  -u "admin:$GRAFANA_PWD" \
  http://collie-dev.org:3000/api/dashboards/db?orgId=1 \
  -d @./grafana/dashboards/workaround-helm.json


