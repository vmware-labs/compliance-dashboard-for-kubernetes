#!/bin/bash
source .env

for i in 1 2
do
	RESP=$(curl -skS -u $COLLIE_ES_USERNAME:$COLLIE_ES_PASSWORD "$COLLIE_ES_URL/collie-k8s-activity-elastic/_search?q=a:9e888726a3908753" | jq ".hits.hits[0]._source")

	if [ "$RESP" = "null" ]; then
    		echo "Not ready yet..."
		sleep 10
	else
    		echo "$RESP"
    		echo "OK"
		exit 0
	fi
done

echo "Failed waiting for agent report in ES"
exit 1
