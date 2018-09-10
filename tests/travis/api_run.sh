#!/bin/bash

set -e

IP=`ip addr s eth0 |grep "inet "|awk '{print $2}' |awk -F "/" '{print $1}'`
echo $IP
docker ps
pybot -v ip:$IP -v HARBOR_PASSWORD:Harbor12345 /home/travis/gopath/src/github.com/goharbor/harbor/tests/robot-cases/Group0-BAT/API.robot
cat /home/travis/gopath/src/github.com/goharbor/harbor/log.html