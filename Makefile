SHELL := /bin/bash

.PHONY: all
all: lint test

.PHONY: test
test:
	go test -failfast ./...

.PHONY: lint
lint:
	golangci-lint run
