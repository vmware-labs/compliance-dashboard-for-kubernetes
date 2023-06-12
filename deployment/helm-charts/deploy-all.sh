#!/bin/bash

source ./delete-agent.sh
source ./delete-server.sh
source ./check-req.sh

echo --------------------------
echo ! Manual operation needed: 
echo Update /etc/hosts, add 'collie.local' to local host public IP.
echo --------------------------

read -p "Press ENTER after the entry has been added..."

ping -c 1 collie.local
if [ $? -ne 0 ]; then
    echo "collie.local is not set to local public IP. Update it in /etc/host"
fi

kubectl create namespace collie-server 

source deploy-es.sh
source deploy-grafana.sh
source deploy-api-server.sh

echo
echo Setup complete.
echo Open browser: http://collie.local:8080/collir/portal/login


