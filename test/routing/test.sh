#!/bin/bash

. $(dirname $0)/../shared/check.sh

set -eux

SCRIPT_DIR=$(cd $(dirname $0); pwd)
cd $SCRIPT_DIR
MONITOR_FILE=/tmp/routing-monitor


if [ -f $MONITOR_FILE ]; then
    rm $MONITOR_FILE
fi

docker-compose up -d

docker-compose exec -T client ip mptcp monitor > $MONITOR_FILE &
sleep 1
docker-compose exec client iperf3 -c localhost -p 5555

docker-compose down

check $MONITOR_FILE 'CREATED' 2
check $MONITOR_FILE ' ESTABLISHED' 2
check $MONITOR_FILE 'ANNOUNCED' 2
check $MONITOR_FILE 'SF_ESTABLISHED' 2
check $MONITOR_FILE 'CLOSED' 1

rm $MONITOR_FILE
