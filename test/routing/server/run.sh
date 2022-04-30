#!/bin/bash

ip mptcp limits set subflow 1
ip mptcp endpoint add 10.123.201.2 dev eth1 signal

# ip route add 10.123.202.0/24 via 10.123.200.3 dev eth0 src 10.123.200.2
# ip route add 10.123.202.0/24 via 10.123.201.3 dev eth1 src 10.123.201.2
ip route add 10.123.202.0/24 nexthop via 10.123.200.3 weight 1 nexthop via 10.123.201.3 weight 1

iperf3 -s &
/mptcp-proxy/mptcp-proxy -m server -p 4444 -r localhost:5201