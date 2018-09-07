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
curl -L https://github.com/docker/compose/releases/download/${DOCKER_COMPOSE_VERSION}/docker-compose-`uname -s`-`uname -m` > docker-compose
chmod +x docker-compose
sudo mv docker-compose /usr/local/bin
IP=`ip addr s eth0 |grep "inet "|awk '{print $2}' |awk -F "/" '{print $1}'`
sudo sed -i '$a DOCKER_OPTS=\"--insecure-registry '$IP':5000\"' /etc/default/docker
sudo service docker restart
sudo service postgresql stop