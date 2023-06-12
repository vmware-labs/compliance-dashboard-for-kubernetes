#!/bin/bash

ping -c 1 collie.local > /dev/null 2>&1

if [ $? -ne 0 ]; then
	echo --------------------------
	echo Manual operation needed:
	echo 1. Use ifconfig to identify local machine public IP.
	echo 2. Update /etc/hosts, add 'collie.local' to local host public IP.
	echo --------------------------
	echo

	read -p "Press ENTER after the entry has been added..."
	echo

	ping -c 1 collie.local > /dev/null 2>&1
	if [ $? -ne 0 ]; then
		echo "Error: collie.local is not set to local public IP. Update it in /etc/host"
		exit 1
	fi
fi
