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