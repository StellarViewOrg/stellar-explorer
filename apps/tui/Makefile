.PHONY: build install run run-rpc run-once run-hybrid run-hybrid-once test test-integration test-all fmt

BINARY_NAME ?= stellar-tui
VERSION ?= 0.1.0
LDFLAGS ?= -s -w -X github.com/miguelnietoa/stellar-explorer/tui/internal/config.Version=$(VERSION)

build:
	go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY_NAME) ./cmd/tui
	@ln -sf $(BINARY_NAME) bin/tui

install:
	go build -ldflags "$(LDFLAGS)" -o "$(shell go env GOPATH)/bin/$(BINARY_NAME)" ./cmd/tui

run: build
	./bin/$(BINARY_NAME)

run-rpc: run

run-once: build
	./bin/$(BINARY_NAME) -no-interactive

run-hybrid: build
	./bin/$(BINARY_NAME) -config config/hybrid.testnet.json

run-hybrid-once: build
	./bin/$(BINARY_NAME) -config config/hybrid.testnet.json -no-interactive

test:
	go test ./...

test-integration:
	go test -tags=integration ./...

test-all: test test-integration

fmt:
	gofmt -w .