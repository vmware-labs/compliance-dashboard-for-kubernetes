

Mac
Install homebrew: https://brew.sh/
Install kubectl: https://formulae.brew.sh/formula/kubernetes-cli
Install minikube: https://minikube.sigs.k8s.io/docs/start/
Install helm chart: https://helm.sh/docs/intro/quickstart/

Mac
Install/upgrade kubectl:
    brew upgrade kubectl
    brew link --overwrite kubernetes-cli

Install/upgrade minikube:
    brew unlink minikube
    brew install minikube
    brew link minikube

Install/upgrade helm:
    brew install helm

Update helm repo:
    helm repo add grafana https://grafana.github.io/helm-charts
    helm repo add elastic https://helm.elastic.co

minikube addons enable default-storageclass
minikube addons enable storage-provisioner

Install ES
    helm install -n collie-server es elastic/elasticsearch -f ./es/values.yaml --wait

Get ES password
    export ES_PASSWORD=$(kubectl get -n collie-server secret/elasticsearch-master-credentials -o jsonpath="{.data.password}" | base64 -d)
    ES_USER=$(kubectl get -n collie-server secret/elasticsearch-master-credentials -o jsonpath="{.data.username}" | base64 -d)
    ES_AUTH=$ES_USER:$ES_PASSWORD
    ES_URL=https://collie-dev.org:9200

Create ES index
```
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
```
    
Install Grafana
    envsubst < ./grafana/values.yaml > tmp.yaml

    #helm install grafana grafana/grafana -f ./grafana/values.yaml -n collie-server --set "grafana.datasources.\"datasources.yaml\".datasources[0].secureJsonData.basicAuthPassword=$ES_PASSWORD"

    helm install grafana grafana/grafana -f tmp.yaml -n collie-server --wait
    rm tmp.yaml

Forward grafana and ES
    kubectl port-forward -n collie-server --address 0.0.0.0 services/elasticsearch-master 9200:9200 & \
    kubectl port-forward -n collie-server --address 0.0.0.0 services/grafana 3000:3000 & \

