#!/bin/sh

set +e

BRANCH=$(git rev-parse --abbrev-ref HEAD)
SWAGGERFILE=https://raw.githubusercontent.com/vmware/harbor/$BRANCH/docs/swagger.yaml
CHERKER=http://online.swagger.io/validator/debug?url=$SWAGGERFILE

echo $SWAGGERFILE

TIMEOUT=5
while [ $TIMEOUT -gt 0 ]; do
    STATUS=$(curl --insecure -s -o /dev/null -w '%{http_code}' $CHERKER)
    if [ $STATUS -eq 200 ]; then
		break
    fi
    TIMEOUT=$(($TIMEOUT - 1))
    sleep 2
done

if [ $TIMEOUT -eq 0 ]; then
    echo "Swagger online checker cannot reach success, but not fail travis."
    exit 0
fi

curl -X GET $CHERKER | grep "{}"  > /dev/null
if [ $? -eq 0 ]; then 
	echo "Swagger yaml check success."
else
	echo "Swagger yaml check fail."
	echo $(curl -X GET $CHERKER)
	exit 1
fi
 