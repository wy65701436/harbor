#!/bin/bash

set -e

sudo apt-get update && sudo apt-get install -y python-dev openjdk-7-jdk
sudo wget https://bootstrap.pypa.io/get-pip.py && python ./get-pip.py && pip install pyasn1 google-apitools==0.5.15 gsutil robotframework robotframework-sshlibrary robotframework-httplibrary requests dbbot robotframework-selenium2library robotframework-pabot --upgrade
sudo make install GOBUILDIMAGE=golang:1.9.2 COMPILETAG=compile_golangimage CLARITYIMAGE=goharbor/harbor-clarity-ui-builder:1.6.0 NOTARYFLAG=true CLAIRFLAG=true