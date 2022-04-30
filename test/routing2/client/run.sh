#!/bin/bash

ip mptcp limits set subflow 1 add_addr_accepted 1
ip mptcp endpoint add 10.123.202.3 dev eth1 subflow

ip route add 10.123.200.0/24 via 10.123.201.2 src 10.123.201.3 metric 10
ip route add 10.123.200.0/24 via 10.123.202.2 src 10.123.202.3 metric 11

#ip route del default via 10.123.201.1 dev eth0
#
#echo "100 rt_eth0" >> /etc/iproute2/rt_tables
#ip rule add from 10.123.201.3/32 table rt_eth0 priority 100
#ip route add table rt_eth0 10.123.201.0/24  dev eth0 proto kernel scope link src 10.123.201.3
#ip route add table rt_eth0 10.123.200.0/24 dev eth0 via 10.123.201.2
#
#echo "200 rt_eth1" >> /etc/iproute2/rt_tables
#ip rule add from 10.123.202.3/32 table rt_eth1 priority 101
#ip route add table rt_eth1 10.123.202.0/24 dev eth1 proto kernel scope link src 10.123.202.3
#ip route add table rt_eth1 10.123.200.0/24 dev eth1 via 10.123.202.2

/mptcp-proxy/mptcp-proxy -m client -p 5555 -r 10.123.200.2:4444