#!/bin/bash

ping -c 1 collie-dev.org > /dev/null 2>&1

if [ $? -ne 0 ]; then
	echo "Error: collie-dev.org is not set to local public IP."
	echo
	echo --------------------------
	echo Manual operation needed:
	echo 1. Use ifconfig to identify local machine public IP.
	echo 2. Update /etc/hosts, add 'collie-dev.org' to local host public IP.
	echo --------------------------
	echo

	exit 1
fi
