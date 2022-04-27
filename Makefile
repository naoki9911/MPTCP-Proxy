.DEFAULT: all

all: build

build:
	go build ./cmd/mptcp-proxy

clean:
	rm mptcp-proxy

.PHONY: all build
