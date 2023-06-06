#!/bin/bash

source .env

GITLAB_URL="https://gitlab.eng.vmware.com"
SCOPE="read_user"

# Request an OAuth access token from GitLab
response=$(curl --silent --request POST \
  --url "$GITLAB_URL/oauth/token" \
  --header 'Content-Type: application/json' \
  --data @- <<EOF
{
  "grant_type": "client_credentials",
  "client_id": "$OAUTH_GITLAB_CLIENTID",
  "client_secret": "$OAUTH_GITLAB_CLIENTSECRET",
  "scope": "$SCOPE"
}
EOF
)

echo $response

# Extract the access token from the response using jq
# access_token=$(echo "$response" | jq -r '.access_token')

# Print the access token
# echo "OAuth access token: $access_token"