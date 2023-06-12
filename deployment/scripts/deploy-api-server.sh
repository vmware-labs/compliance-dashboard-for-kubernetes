#!/bin/bash

B64CMD="base64"

source .env

# Delete deployments
kubectl delete namespace collie-server collie-agent --wait=true

# redeploy
export ES_KEY_B64=$(echo -n $ES_KEY | $B64CMD)
export OAUTH_CSP_CLIENTID_B64=$(echo -n $OAUTH_CSP_CLIENTID | $B64CMD)
export OAUTH_CSP_CLIENTSECRET_B64=$(echo -n $OAUTH_CSP_CLIENTSECRET | $B64CMD)
export OAUTH_GITLAB_CLIENTID_B64=$(echo -n $OAUTH_GITLAB_CLIENTID | $B64CMD)
export OAUTH_GITLAB_CLIENTSECRET_B64=$(echo -n $OAUTH_GITLAB_CLIENTSECRET | $B64CMD)

envsubst < api-server.yaml | kubectl apply -f -
kubectl wait deployment -n collie-server api-server --for condition=Available=True --timeout=90s
AUTH_TOKEN=gitlab/$(source auth-gitlab.sh | jq -r '.access_token')
sleep 10
curl -skH "Authorization: $AUTH_TOKEN" https://collie.eng.vmware.com/collie/api/v1/onboarding/bootstrap

