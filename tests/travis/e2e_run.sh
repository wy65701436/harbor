#!/bin/bash

set -e

IP=`ip addr s eth0 |grep "inet "|awk '{print $2}' |awk -F "/" '{print $1}'`
echo $IP
sleep 20
docker ps
pybot -v ip:$IP -v HARBOR_PASSWORD:Harbor12345 /home/travis/gopath/src/github.com/goharbor/harbor/tests/robot-cases/Group0-BAT/E2E.robot
cat ./log.html