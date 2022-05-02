#!/bin/bash

#iptables -t mangle -F
#iptables -t mangle -N DIVERT
#iptables -t mangle -A DIVERT -j MARK --set-mark 1
#iptables -t mangle -A DIVERT -j ACCEPT
#iptables -t mangle -A PREROUTING -p tcp -j DIVERT
#iptables -t mangle -A PREROUTING -p tcp --dport 8080 -j TPROXY \
#  --tproxy-mark 0x1/0x1 --on-port 52000
#
#ip rule add fwmark 1 lookup 100
#ip route add local 0.0.0.0/0 dev lo table 100


sudo iptables -t nat -A OUTPUT -p tcp -m owner ! --uid-owner mptcpproxy --dport 1194 -j REDIRECT --to-port 52000
