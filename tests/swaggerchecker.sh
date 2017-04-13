#!/bin/sh

set +e

echo $TRAVIS_PULL_REQUEST_SLUG
echo $TRAVIS_PULL_REQUEST_SHA

if [ -z "$1" ]; then
	echo '* Required input `git repo name` not provided!'
	exit 1
fi

if [ -z "$2" ]; then
	echo '* Required input `git commit id` not provided!'
	exit 1
fi

SWAGGER_ONLINE_VALIDATOR="http://online.swagger.io/validator"
HARBOR_SWAGGER_FILE="https://raw.githubusercontent.com/$1/$2/docs/swagger.yaml"
HARBOR_SWAGGER_VALIDATOR_URL="$SWAGGER_ONLINE_VALIDATOR/debug?url=$HARBOR_SWAGGER_FILE"
echo $HARBOR_SWAGGER_VALIDATOR_URL

# Now try to ping swagger online validator, then to use it to do the validation.
eval curl -f -I $SWAGGER_ONLINE_VALIDATOR
curl_ping_res=$?
if [ ${curl_ping_res} -eq 0 ]; then
	echo "* cURL ping swagger validator returned success"
else
	echo "* cURL ping swagger validator returned an error (${curl_ping_res}), but don't fail the travis CI here."
	exit 0
fi

# Use the swagger online validator to validate the harbor swagger file.
eval curl -s $HARBOR_SWAGGER_VALIDATOR_URL > output.json
curl_validate_res=$?
validate_expected_results="{}"
validate_actual_results=$(cat < output.json)

if [ ${curl_validate_res} -eq 0 ]; then
	if [ $validate_actual_results = $validate_expected_results ]; then
		echo "* cURL check Harbor swagger file returned success"
	else
		echo "* cURL check Harbor swagger file returned an error ($validate_actual_results)"
	fi
else
	echo "* cURL check Harbor swagger file returned an error (${curl_validate_res})"
	exit ${curl_validate_res}
fi
