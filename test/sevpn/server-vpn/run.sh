#!/bin/bash

VPNCMD=/usr/vpnserver/vpncmd
VPNSERVER=/usr/vpnserver/vpnserver

ip route del default
ip link add br0 type bridge
ip link set eth1 master br0

$VPNSERVER start

sleep 1

ip link set tap_default master br0
ip link set up br0

while true
do
  sleep 1
done