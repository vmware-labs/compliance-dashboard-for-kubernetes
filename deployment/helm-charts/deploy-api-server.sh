#!/bin/bash

helm list -n collie-server -a | grep api-server > /dev/null 2>&1
if [ $? -eq 0 ]; then
	helm uninstall -n collie-server api-server --wait
fi

helm install -n collie-server api-server ./api-server --wait

sleep 10

kubectl port-forward -n collie-server --address 0.0.0.0 services/api-server 8080:8080

