#!/bin/bash

helm list -n collie-server -a | grep es > /dev/null 2>&1
if [ $? -eq 0 ]; then
  helm uninstall -n collie-server es --wait
fi

echo Deploy ES...
helm install -n collie-server es elastic/elasticsearch -f ./es/values.yaml -f ./es/values-custom.yaml --wait

sleep 10

export ES_PASSWORD=$(kubectl get -n collie-server secret/elasticsearch-master-credentials -o jsonpath="{.data.password}" | base64 -d)

ES_USER=$(kubectl get -n collie-server secret/elasticsearch-master-credentials -o jsonpath="{.data.username}" | base64 -d)

ES_AUTH=$ES_USER:$ES_PASSWORD
ES_URL=https://collie.local:9200

echo ES_AUTH: $ES_AUTH
Echo Forwarding ES in background...
kubectl port-forward -n collie-server --address 0.0.0.0 services/elasticsearch-master 9200:9200 &

sleep 5

Echo Clean up ES data...
curl -k -u $ES_AUTH -X DELETE $ES_URL/collie-k8s-elastic
Echo Recreate ES index...
curl -k -u $ES_AUTH -X PUT -H "Content-Type: application/json" -d '
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
}' $ES_URL/collie-k8s-elastic

echo

