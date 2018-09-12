#!/bin/bash

set -e

export POSTGRESQL_HOST=$1
export REGISTRY_URL=$1:5000

sudo make run_clarity_ut CLARITYIMAGE=goharbor/harbor-clarity-ui-builder:${UI_BUILDER_VERSION}
cat ./src/ui_ng/npm-ut-test-results
sudo docker-compose -f ./make/docker-compose.test.yml up -d
sleep 10
sudo make go_check
./tests/pushimage.sh
docker ps

go test -race -i ./src/ui ./src/adminserver ./src/jobservice
sudo -E env "PATH=$PATH" "POSTGRES_MIGRATION_SCRIPTS_PATH=/home/travis/gopath/src/github.com/goharbor/harbor/make/migrations/postgresql/" ./tests/coverage4gotest.sh
goveralls -coverprofile=profile.cov -service=travis-ci