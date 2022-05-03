.DEFAULT: all

all: build

build:
	go build ./cmd/mptcp-proxy

build-static:
	go build -a -tags netgo -installsuffix netgo --ldflags '-extldflags "-static"' ./cmd/mptcp-proxy

install: build 
	sudo install mptcp-proxy /usr/local/bin
	sudo setcap cap_net_admin+ep /usr/local/bin/mptcp-proxy

test:
	test/simple/test.sh
	test/routing/test.sh
	test/routing2/test.sh
	test/sevpn/test.sh

clean:
	rm mptcp-proxy

.PHONY: all build build-static install test
