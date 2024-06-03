all: test vet fmt

test:
	go test -v ./...
.PHONY: test

vet:
	go vet ./...
.PHONY: vet

fmt:
	go fmt ./...
.PHONY: fmt
