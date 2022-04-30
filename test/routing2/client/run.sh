#!/bin/bash

ip mptcp limits set subflow 1 add_addr_accepted 1
ip mptcp endpoint add 10.123.202.3 dev eth1 subflow

ip route add 10.123.200.0/24 via 10.123.201.2 metric 10
ip route add 10.123.200.0/24 via 10.123.202.2 metric 20


/mptcp-proxy/mptcp-proxy -m client -p 5555 -r 10.123.200.2:4444