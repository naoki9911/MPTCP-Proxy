.DEFAULT: all

all: build

build:
	go build ./cmd/mptcp-proxy

test:
	test/simple/test.sh
	test/routing/test.sh
	test/routing2/test.sh

clean:
	rm mptcp-proxy

.PHONY: all build test
