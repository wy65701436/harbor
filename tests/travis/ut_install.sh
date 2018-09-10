#!/bin/bash

set -e

sudo apt-get update && sudo apt-get install -y libldap2-dev
go get -d github.com/docker/distribution
go get -d github.com/docker/libtrust
go get -d github.com/lib/pq
go get github.com/golang/lint/golint
go get github.com/GeertJohan/fgt
go get github.com/dghubble/sling
go get github.com/stretchr/testify
go get golang.org/x/tools/cmd/cover
go get github.com/mattn/goveralls
go get -u github.com/client9/misspell/cmd/misspell
sudo service postgresql stop

IP=`ip addr s eth0 |grep "inet "|awk '{print $2}' |awk -F "/" '{print $1}'`
export POSTGRESQL_HOST=$IP
export REGISTRY_URL=$IP:5000

cd tests && sudo ./ldapprepare.sh && sudo ./admiral.sh && cd ..
sudo make compile_adminserver
sudo make -f make/photon/Makefile _build_adminserver _build_db _build_registry -e VERSIONTAG=dev -e CLAIRDBVERSION=dev -e REGISTRYVERSION=${REG_VERSION}
sudo sed -i 's/__reg_version__/${REG_VERSION}-dev/g' ./make/docker-compose.test.yml
sudo sed -i 's/__version__/dev/g' ./make/docker-compose.test.yml
sudo mkdir -p ./make/common/config/registry/ && sudo mv ./tests/reg_config.yml ./make/common/config/registry/config.yml