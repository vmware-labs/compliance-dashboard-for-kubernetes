#!/bin/bash

source .env

GITLAB_URL="https://gitlab.eng.vmware.com"
SCOPES="read_user"

# Request an OAuth access token from GitLab
response=$(curl --silent --request POST \
  --url "$GITLAB_URL/oauth/token" \
  --header 'content-type: application/x-www-form-urlencoded' \
  --data "client_id=$OAUTH_GITLAB_CLIENTID&client_secret=$OAUTH_GITLAB_CLIENTSECRET&grant_type=client_credentials&scope=$SCOPES"
)

echo $response

# Extract the access token from the response using jq
# ACCESS_TOKEN=$(echo "$response" | jq -r '.access_token')
# echo Access token: $ACCESS_TOKEN
# echo Token info: $(curl -s $GITLAB_URL/oauth/token/info?access_token=$ACCESS_TOKEN)

# USER_RESPONSE=$(curl -s $GITLAB_URL/oauth/userinfo?access_token=$ACCESS_TOKEN)
# echo OAuth user info: $USER_RESPONSE

# USER_RESPONSE=$(curl -s $GITLAB_URL/api/v4/user?access_token=$ACCESS_TOKEN)
# echo Gitlab info: $USER_RESPONSE

