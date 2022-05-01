FROM ubuntu:22.04

RUN apt-get update -y && apt-get upgrade -y
RUN apt-get install -y iproute2 golang iperf3 make ca-certificates git tcpdump iputils-ping nginx curl net-tools

RUN mkdir /mptcp-proxy
WORKDIR /mptcp-proxy
COPY . .

RUN make build

CMD [ "/bin/bash", "/run.sh"]