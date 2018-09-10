#!/bin/bash

set -e

sudo apt-get update && sudo apt-get install -y --no-install-recommends python-dev openjdk-7-jdk libssl-dev && sudo apt-get autoremove -y && sudo rm -rf /var/lib/apt/lists/*
sudo wget https://bootstrap.pypa.io/get-pip.py && sudo python ./get-pip.py && sudo pip install --ignore-installed urllib3 chardet requests && sudo pip install robotframework robotframework-httplibrary requests dbbot robotframework-pabot --upgrade
sudo make install GOBUILDIMAGE=golang:1.9.2 COMPILETAG=compile_golangimage CLARITYIMAGE=goharbor/harbor-clarity-ui-builder:1.6.0 NOTARYFLAG=true CLAIRFLAG=true