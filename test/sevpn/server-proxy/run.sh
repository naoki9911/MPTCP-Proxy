#!/bin/bash

ip route del default

ip mptcp limits set subflow 1
ip mptcp endpoint add 10.102.1.2 dev eth1 signal

nginx
iperf3 -s &
/mptcp-proxy/mptcp-proxy -m server -p 4444 -r 10.100.0.3:1194
