#!/bin/bash

NAMESPACE=collie-server
TARGET_CONFIG=pod/elasticsearch-master-0

kubectl -n $NAMESPACE get $TARGET_CONFIG -ojson | jq . > pod-debug1.json 

#jq "del(.spec.containers[0].readinessProbe)" pod-debug.json > pod-debug.json
jq "del(.spec.initContainers[0].command)" pod-debug1.json > pod-debug2.json
jq "del(.spec.initContainers[0].args)" pod-debug2.json > pod-debug3.json
jq '.spec.initContainers[0] += {command: ["/bin/bash"], args: ["-c", "while true; do sleep 30; done;"]}' pod-debug3.json > pod-debug4.json
jq '.metadata.name = "debug-image"' pod-debug4.json > pod-debug5.json

cat pod-debug5.json | yq -P > pod-debug.yml

kubectl -n $NAMESPACE delete pod debug-image 
sleep 20
kubectl -n $NAMESPACE apply -f pod-debug.yml
sleep 20
kubectl -n $NAMESPACE exec -it debug-image -- /bin/bash

# rm -f pod-debug*.json
# rm -f pod-debug*.yml
