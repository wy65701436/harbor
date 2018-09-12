#!/bin/bash

set +e
sudo rm -rf /data
sudo mkdir -p /data

set -e

# prepare cert ...
IP=`ip addr s eth0 |grep "inet "|awk '{print $2}' |awk -F "/" '{print $1}'`
sudo sed "s/127.0.0.1/$IP/" -i tests/generateCerts.sh
sudo ./tests/generateCerts.sh
sudo mkdir -p /etc/docker/certs.d/$IP
sudo cp ./harbor_ca.crt /etc/docker/certs.d/$IP/


sudo ./tests/hostcfg.sh LDAP
cd tests && sudo ./ldapprepare.sh && cd ..
sudo apt-get update && sudo apt-get install -y --no-install-recommends python-dev openjdk-7-jdk libssl-dev && sudo apt-get autoremove -y && sudo rm -rf /var/lib/apt/lists/*
sudo wget https://bootstrap.pypa.io/get-pip.py && sudo python ./get-pip.py && sudo pip install --ignore-installed urllib3 chardet requests && sudo pip install robotframework robotframework-httplibrary requests dbbot robotframework-pabot --upgrade
sudo make swagger_client
sudo make install GOBUILDIMAGE=golang:1.9.2 COMPILETAG=compile_golangimage CLARITYIMAGE=goharbor/harbor-clarity-ui-builder:1.6.0 NOTARYFLAG=true CLAIRFLAG=true CHARTFLAG=true
sleep 10