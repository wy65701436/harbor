#!/bin/bash

set -e

export POSTGRESQL_HOST=$IP
export REGISTRY_URL=$IP:5000

cd tests && sudo ./ldapprepare.sh && cd ..
sudo ./tests/admiral.sh
sudo make compile_adminserver
sudo make -f make/photon/Makefile _build_adminserver _build_db _build_registry -e VERSIONTAG=dev -e CLAIRDBVERSION=dev -e REGISTRYVERSION=${REG_VERSION}
sudo sed -i 's/__reg_version__/${REG_VERSION}-dev/g' ./make/docker-compose.test.yml
sudo sed -i 's/__version__/dev/g' ./make/docker-compose.test.yml
sudo mkdir -p ./make/common/config/registry/ && sudo mv ./tests/reg_config.yml ./make/common/config/registry/config.yml