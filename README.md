# MPTCP-Proxy

MPTCP-Proxy is TCP proxy to provide Multipath TCP for TCP applications.

![The Overview of MPTCP-Proxy](/assets/overview.png) 

# Usage

```
Usage of ./mptcp-proxy:
  -m, --mode string     specify mode (server or client)
  -p, --port int        local bind port
  -r, --remote string   remote address (ex. 127.0.0.1:8080)
```

# Environment

- Ubuntu 22.04 LTS

# Testing Environment with docker-compose
You can test MPTCP-Proxy with docker-compose(`test/docker-compose.yml`)

![Test Environment](/assets/test-env.png)

```console
$ cd test
$ docker-compose up -d
$ docker-compose exec client /bin/bash

root@73640481f772:/mptcp-proxy# iperf3 -c localhost -p 5555
Connecting to host localhost, port 5555
[  5] local 127.0.0.1 port 35886 connected to 127.0.0.1 port 5555
[ ID] Interval           Transfer     Bitrate         Retr  Cwnd
[  5]   0.00-1.00   sec  1.38 GBytes  11.8 Gbits/sec    0   1.81 MBytes       
[  5]   1.00-2.00   sec  1.36 GBytes  11.7 Gbits/sec    0   1.81 MBytes       
[  5]   2.00-3.00   sec  1.34 GBytes  11.5 Gbits/sec    0   1.81 MBytes       
[  5]   3.00-4.00   sec  1.35 GBytes  11.6 Gbits/sec    0   1.81 MBytes       
[  5]   4.00-5.00   sec  1.23 GBytes  10.6 Gbits/sec    0   1.81 MBytes       
[  5]   5.00-6.00   sec  1.18 GBytes  10.2 Gbits/sec    0   1.81 MBytes       
[  5]   6.00-7.00   sec  1.18 GBytes  10.2 Gbits/sec    0   1.81 MBytes       
[  5]   7.00-8.00   sec  1.18 GBytes  10.2 Gbits/sec    0   1.81 MBytes       
[  5]   8.00-9.00   sec  1.23 GBytes  10.6 Gbits/sec    0   1.81 MBytes       
[  5]   9.00-10.00  sec  1.29 GBytes  11.1 Gbits/sec    0   1.81 MBytes       
- - - - - - - - - - - - - - - - - - - - - - - - -
[ ID] Interval           Transfer     Bitrate         Retr
[  5]   0.00-10.00  sec  12.7 GBytes  10.9 Gbits/sec    0             sender
[  5]   0.00-10.00  sec  12.7 GBytes  10.9 Gbits/sec                  receiver

iperf Done.
```

# Example of MPTCP Packtes

![Wireshark](/assets/wireshark.png)