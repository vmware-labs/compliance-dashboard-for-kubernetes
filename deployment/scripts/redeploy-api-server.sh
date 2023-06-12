#!/bin/bash

source .env

if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        B64CMD="base64 --wrap=0"
elif [[ "$OSTYPE" == "darwin"* ]]; then
        B64CMD="base64"
elif [[ "$OSTYPE" == "cygwin" ]]; then
        echo "TODO: OSTYPE cygwin"
	exit 1
elif [[ "$OSTYPE" == "msys" ]]; then
        echo "TODO: OSTYPE msys"
	exit 1
elif [[ "$OSTYPE" == "win32" ]]; then
        echo "TODO: OSTYPE win32"
	exit 1
elif [[ "$OSTYPE" == "freebsd"* ]]; then
        echo "TODO: OSTYPE freebsd"
	exit 1
else
	echo "TODO: unknown OSTYPE "$OSTYPE
	exit 1
fi

# Delete deployments
kubectl delete namespace collie-server collie-agent --wait=true
sleep 30

minikube image rm collie.azurecr.io/collie-api-server:1
minikube image rm collie.azurecr.io/collie-agent:1

# clear data
source es-recreate-index.sh

# redeploy
export ES_KEY_B64=$(echo -n $ES_KEY | base64 --wrap=0)
export OAUTH_CSP_CLIENTID_B64=$(echo -n $OAUTH_CSP_CLIENTID | $B64CMD)
export OAUTH_CSP_CLIENTSECRET_B64=$(echo -n $OAUTH_CSP_CLIENTSECRET | $B64CMD)
export OAUTH_GITLAB_CLIENTID_B64=$(echo -n $OAUTH_GITLAB_CLIENTID | $B64CMD)
export OAUTH_GITLAB_CLIENTSECRET_B64=$(echo -n $OAUTH_GITLAB_CLIENTSECRET | $B64CMD)

envsubst < api-server.yaml | kubectl apply -f -
kubectl wait deployment -n collie-server api-server --for condition=Available=True --timeout=90s
AUTH_TOKEN=gitlab/$(source auth-gitlab.sh | jq -r '.access_token')
sleep 10
BOOTSTRAP_CMD=$(curl -skH "Authorization: $AUTH_TOKEN" https://collie.eng.vmware.com/collie/api/v1/onboarding/bootstrap | jq -r ".cmd")

AGENT_ID=$(echo $BOOTSTRAP_CMD | sed -n 's/.*aid\=\(.*\)\".*/\1/p')

echo -e '#!/bin/bash\n' > deploy-agent.sh
echo $BOOTSTRAP_CMD >> deploy-agent.sh
chmod +x deploy-agent.sh

source ./deploy-agent.sh
sleep 30
kubectl -n collie-agent logs deployment/agent

# wait for data appear in ES

echo $RESP

for i in 1 2
do
        RESP=$(curl -skS -u $COLLIE_ES_USERNAME:$COLLIE_ES_PASSWORD "$COLLIE_ES_URL/collie-k8s-elastic/_search?q=a:$AGENT_ID" | jq ".hits.hits[0]._source")

        if [ "$RESP" = "null" ]; then
                echo "Not ready yet..."
                sleep 10
        else
                echo "$RESP"
                echo "OK"
		kubectl delete namespace collie-agent --wait=true
		exit 0
        fi
done

echo "Failed waiting for agent report in ES"
exit 1
