.PHONY: default lint test

export GO111MODULE=on

default: lint test

lint:
	golangci-lint run

test:
	go test ./...
