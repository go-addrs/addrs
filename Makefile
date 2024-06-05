all: build test vet fmt

build:
	go build ./...
.PHONY: build

test:
	go test -c ./ip
	go test -c ./ipv4
	go test -c ./ipv6
	go test -v ./...
.PHONY: test

vet:
	go vet ./...
.PHONY: vet

fmt:
	go fmt ./...
.PHONY: fmt
