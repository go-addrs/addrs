all: test vet staticcheck fmt

test:
	go test -v ./...
.PHONY: test

vet:
	go vet ./...
.PHONY: vet

staticcheck:
	staticcheck ./...
.PHONY: staticcheck

fmt:
	go fmt ./...
.PHONY: fmt
