#!/bin/bash

source ./delete-agent.sh

helm list -n collie-server -a | grep api-server > /dev/null 2>&1
if [ $? -eq 0 ]; then
  helm uninstall -n collie-server api-server --wait
fi
helm list -n collie-server -a | grep grafana > /dev/null 2>&1
if [ $? -eq 0 ]; then
  helm uninstall -n collie-server grafana --wait
fi
helm list -n collie-server -a | grep es > /dev/null 2>&1
if [ $? -eq 0 ]; then
  helm uninstall -n collie-server es --wait
fi

kubectl get ns/collie-server > /dev/null 2>&1
if [ $? -eq 0 ]; then
	kubectl delete namespace collie-server --wait
fi


