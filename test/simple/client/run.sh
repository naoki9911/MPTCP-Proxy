#!/bin/bash

ip mptcp limits set subflow 1 add_addr_accepted 1

/mptcp-proxy/mptcp-proxy -m client -p 5555 -r 10.123.200.2:4444