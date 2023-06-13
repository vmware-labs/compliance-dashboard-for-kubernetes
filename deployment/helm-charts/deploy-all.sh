#!/bin/bash

source ./delete-all.sh
source ./check-req.sh

kubectl create namespace collie-server 

source deploy-es.sh
source deploy-grafana.sh
source deploy-server.sh

echo
echo Complete.


