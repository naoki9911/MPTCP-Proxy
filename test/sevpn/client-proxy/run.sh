#!/bin/bash

ip route del default

ip mptcp limits set subflow 1 add_addr_accepted 1

/mptcp-proxy/mptcp-proxy -m client -p 4444 -r 10.102.0.2:4444
