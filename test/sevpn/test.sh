#!/bin/bash

. $(dirname $0)/../shared/check.sh

set -eux

SCRIPT_DIR=$(cd $(dirname $0); pwd)
cd $SCRIPT_DIR
MONITOR_FILE=/tmp/sevpn-monitor


if [ -f $MONITOR_FILE ]; then
    rm $MONITOR_FILE
fi

docker-compose build
docker-compose up -d

docker-compose exec -T client-proxy ip mptcp monitor > $MONITOR_FILE &
sleep 2
docker-compose exec client-vpn ping 10.100.1.3 -c 5

docker-compose down

check $MONITOR_FILE 'CREATED' 2
check $MONITOR_FILE ' ESTABLISHED' 2
check $MONITOR_FILE 'ANNOUNCED' 2
check $MONITOR_FILE 'SF_ESTABLISHED' 2

# some connections are not tracked?
#check $MONITOR_FILE 'CLOSED' 3

rm $MONITOR_FILE
