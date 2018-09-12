#!/bin/bash

set +e

IP=`ip addr s eth0 |grep "inet "|awk '{print $2}' |awk -F "/" '{print $1}'`
docker ps

# run db auth api cases
if [ "$1" = 'DB' ]; then
    pybot -v ip:$IP -v HARBOR_PASSWORD:Harbor12345 /home/travis/gopath/src/github.com/goharbor/harbor/tests/robot-cases/Group0-BAT/API_DB.robot
fi
# run ldap api cases
if [ "$1" = 'LDAP' ]; then
    pybot -v ip:$IP -v HARBOR_PASSWORD:Harbor12345 /home/travis/gopath/src/github.com/goharbor/harbor/tests/robot-cases/Group0-BAT/API_LDAP.robot
fi

#cat /home/travis/gopath/src/github.com/goharbor/harbor/log.html
ls -la /var/log/harbor
sudo cat /var/log/harbor/ui.log
