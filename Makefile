.DEFAULT: all

all: build

build:
	go build ./cmd/mptcp-proxy

install: build 
	sudo install mptcp-proxy /usr/local/bin
	sudo setcap cap_net_admin+ep /usr/local/bin/mptcp-proxy

test:
	test/simple/test.sh
	test/routing/test.sh
	test/routing2/test.sh

clean:
	rm mptcp-proxy

.PHONY: all build install test
