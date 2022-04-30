#!/bin/bash

ip mptcp limits set subflow 1
#ip mptcp endpoint add 10.123.201.2 dev eth1 signal

ip route add 10.123.201.0/24 via 10.123.200.3
ip route add 10.123.202.0/24 via 10.123.200.3

iperf3 -s &
/mptcp-proxy/mptcp-proxy -m server -p 4444 -r localhost:5201