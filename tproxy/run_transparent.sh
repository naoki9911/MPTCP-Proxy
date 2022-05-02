#!/bin/bash

SCRIPT_DIR=$(cd $(dirname $0); pwd)

sudo -u mptcpproxy bash -c "/usr/local/bin/mptcp-proxy  -m client -p 52000 -t"