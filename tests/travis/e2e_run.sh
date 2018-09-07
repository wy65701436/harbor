#!/bin/bash

set -e

IP=`ip addr s eth0 |grep "inet "|awk '{print $2}' |awk -F "/" '{print $1}'`

pwd

docker run -i --privileged -v /harbor/workspace/harbor_nightly_executor/test-case:/travis -v /harbor/ca:/ca -w /drone vmware/harbor-e2e-engine:1.38 pybot -v ip:$IP -v notaryServerEndpoint:$IP:4443 -v ip1: -v HARBOR_PASSWORD:Harbor12345 /travis/tests/robot-cases/Group0-BAT/BAT.robot