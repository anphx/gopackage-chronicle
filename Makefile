.PHONY: help run-server run-indexer test lint build migrate-up migrate-down

help:
	@echo "Available targets:"
	@echo "  run-server      - Run the API server"
	@echo "  run-indexer     - Run the indexer once manually"
	@echo "  test            - Run all Go tests"
	@echo "  lint            - Run golangci-lint"
	@echo "  build           - Build both binaries"
	@echo "  migrate-up      - Apply pending migrations"
	@echo "  migrate-down    - Roll back the last migration"

run-server:
	cd backend && go run ./cmd/server

run-indexer:
	cd backend && go run ./cmd/indexer

test:
	cd backend && go test -v -race ./...

lint:
	cd backend && golangci-lint run

build:
	cd backend && go build -o bin/server ./cmd/server
	cd backend && go build -o bin/indexer ./cmd/indexer
