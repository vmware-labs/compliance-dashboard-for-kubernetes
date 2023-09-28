#!/bin/bash

helm list -n collie-server -a | grep api-server > /dev/null 2>&1
if [ $? -eq 0 ]; then
	helm uninstall -n collie-server api-server --wait
fi

cd ../../helm-charts
helm install -n collie-server api-server ./api-server --wait

sleep 10

#source ./kill-processes-by-keyword.sh "kubectl port-forward -n collie-server --address 0.0.0.0 services/api-server"
#kubectl port-forward -n collie-server --address 0.0.0.0 services/api-server 8080:8080 &

sleep 2

echo
echo Open browser: http://collie-dev.org:8080/collie/portal/login
echo

