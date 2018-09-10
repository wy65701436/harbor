#!/bin/bash

set -e

IP=`ip addr s eth0 |grep "inet "|awk '{print $2}' |awk -F "/" '{print $1}'`
echo $IP
docker ps
docker run -i --privileged -v /home/travis/gopath/src/github.com/goharbor/harbor:/drone -v /harbor/ca:/ca -w /drone vmware/harbor-e2e-engine:1.41 pybot -v ip:$IP -v notaryServerEndpoint:$IP:4443 -v ip1: -v HARBOR_PASSWORD:Harbor12345 /drone/tests/robot-cases/Group0-BAT/BAT.robot

cat ./log.html