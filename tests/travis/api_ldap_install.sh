#!/bin/bash

set +e
rm -rf /data
mkdir -p /data

set -e
sudo ./tests/generateCerts.sh
sudo ./tests/hostcfg.sh
sudo sed "s/db_auth/ldap_auth/" -i make/harbor.cfg
sudo sed "s/ldaps://ldap.mydomain.com/$IP/" -i make/harbor.cfg
sudo sed "s/#ldap_searchdn = uid=searchuser,ou=people,dc=mydomain,dc=com/ldap_searchdn = cn=admin,dc=example,dc=com/" -i make/harbor.cfg
sudo sed "s/#ldap_search_pwd = password/ldap_search_pwd = admin/" -i make/harbor.cfg
sudo sed "s/ldap_basedn = ou=people,dc=mydomain,dc=com/ldap_basedn = dc=example,dc=com/" -i make/harbor.cfg
sudo sed "s/#ldap_filter = (objectClass=person)/ldap_filter = (&(objectclass=inetorgperson)(memberof=cn=harbor_users,ou=groups,dc=example,dc=com))/" -i make/harbor.cfg
sudo sed "s/ldap_uid = uid/ldap_uid = cn/" -i make/harbor.cfg
sudo apt-get update && sudo apt-get install -y --no-install-recommends python-dev openjdk-7-jdk libssl-dev && sudo apt-get autoremove -y && sudo rm -rf /var/lib/apt/lists/*
sudo wget https://bootstrap.pypa.io/get-pip.py && sudo python ./get-pip.py && sudo pip install --ignore-installed urllib3 chardet requests && sudo pip install robotframework robotframework-httplibrary requests dbbot robotframework-pabot --upgrade
sudo make swagger_client
sudo make install GOBUILDIMAGE=golang:1.9.2 COMPILETAG=compile_golangimage CLARITYIMAGE=goharbor/harbor-clarity-ui-builder:1.6.0 NOTARYFLAG=true CLAIRFLAG=true CHARTFLAG=true
sleep 10
cd tests && sudo ./ldapprepare.sh && cd ..