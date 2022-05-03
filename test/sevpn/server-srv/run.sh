#!/bin/bash

ip route del default

nginx
iperf3 -s &

while true
do
  sleep 1
done