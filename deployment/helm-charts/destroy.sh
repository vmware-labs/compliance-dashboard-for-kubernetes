#!/bin/bash

helm delete -n collie-server api-server --wait
helm delete -n collie-server grafana --wait
helm delete -n collie-server es --wait

kubectl delete namespace collie-agent --wait
kubectl delete namespace collie-server --wait
