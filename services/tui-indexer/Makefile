.PHONY: build test test-all lint fmt run-live run-backfill run-serve migrate clean

build:
	go build -o bin/tui-indexer ./cmd/tui-indexer

test:
	go test ./... -short

test-all:
	go test ./...

fmt:
	gofmt -w .

lint:
	go vet ./...

run-live: build
	./bin/tui-indexer live

run-backfill: build
	./bin/tui-indexer backfill

run-serve: build
	./bin/tui-indexer serve

migrate: build
	./bin/tui-indexer migrate

clean:
	rm -rf bin/
