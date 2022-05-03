#!/bin/bash

VPNCMD=/usr/vpnserver/vpncmd
VPNCLIENT=/usr/vpnclient/vpnclient

ip route del default

$VPNCLIENT start

sleep 1

$VPNCMD localhost /CLIENT /CMD AccountConnect ConTest
ip a add 10.100.1.5/24 dev vpn_test

while true
do
  sleep 1
done