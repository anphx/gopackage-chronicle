.PHONY: help run-server run-indexer test lint build migrate-up migrate-down docker-up-dev docker-up-prod docker-down

help:
	@echo "Available targets:"
	@echo "  run-server      - Run the API server"
	@echo "  run-indexer     - Run the indexer once manually"
	@echo "  test            - Run all Go tests"
	@echo "  lint            - Run golangci-lint"
	@echo "  build           - Build both binaries"
	@echo "  docker-up-dev   - Start containers with dev env (local postgres)"
	@echo "  docker-up-prod  - Start containers with prod env (Supabase)"
	@echo "  docker-down     - Stop and remove containers"

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

docker-up-dev:
	docker compose --env-file .env.dev up -d --build

docker-up-prod:
	docker compose --env-file .env.prod up -d --build

docker-down:
	docker compose down