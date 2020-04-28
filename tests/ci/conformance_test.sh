set -e

harbor_logs_bucket="harbor-ci-logs"

# GS util
function uploader {
   sudo gsutil cp $1 gs://$2/$1
   sudo gsutil acl ch -u AllUsers:R gs://$2/$1
}

echo "get the conformance testing code..."
## ToDo use the official code as PR https://github.com/opencontainers/distribution-spec/pull/144 merged
git clone -b disable-cookie https://www.github.com/wy65701436/distribution-spec.git

echo "create testing project"
STATUS=$(curl -w '%{http_code}' -H 'Content-Type: application/json' -H 'Accept: application/json' -X POST -u "admin:Harbor12345" -s --insecure "https://$IP/api/v2.0/projects" --data '{"project_name":"conformance","metadata":{"public":"false"},"storage_limit":-1}')
if [ $STATUS -ne 201 ]; then
		exit 1
fi

echo "run conformance test..."
export OCI_ROOT_URL="https://$1"
export OCI_NAMESPACE="conformance/testrepo"
export OCI_USERNAME="admin"
export OCI_PASSWORD="Harbor12345"
export OCI_DEBUG="true"
## will add more test, so far only cover pull & push
export OCI_TEST_PUSH=1
export OCI_TEST_PULL=1

cd ./distribution-spec/conformance
go test .

uploader report.html $harbor_logs_bucket
