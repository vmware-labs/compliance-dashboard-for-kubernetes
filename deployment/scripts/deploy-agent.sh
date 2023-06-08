#!/bin/bash

curl -skH "Authorization: Bearer elastic:16b9f5de2c23718edbf713731584fbb3" "https://collie.eng.vmware.com/collie/api/v1/onboarding/agent.yaml?provider=AKS&aid=a63aefd2af7028be" | kubectl apply -f -
