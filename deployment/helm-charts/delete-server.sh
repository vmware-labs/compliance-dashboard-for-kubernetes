#!/bin/bash

kubectl get ns/collie-server > /dev/null 2>&1
if [ $? -eq 0 ]; then
	kubectl delete namespace collie-server --wait
fi
