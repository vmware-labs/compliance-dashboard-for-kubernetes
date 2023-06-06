#!/bin/bash

curl -skH "Authorization: Token f7340bab6c7ce2b213e90040902993f0" "https://collie.eng.vmware.com/collie/api/v1/onboarding/agent.yaml?provider=AKS&aid=24f74f798f0884c7" | kubectl apply -f -
