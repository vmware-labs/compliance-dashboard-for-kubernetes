#!/bin/bash

kubectl get ns/collie-agent > /dev/null 2>&1
if [ $? -eq 0 ]; then
	kubectl delete namespace collie-agent --wait
fi
