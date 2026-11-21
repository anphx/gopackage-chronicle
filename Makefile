.PHONY: help db-up run-server run-indexer test lint build docker-up-dev docker-up-prod docker-down

# Env file loaded by run-server / run-indexer. Override with ENV_FILE=.env.prod.
ENV_FILE ?= .env.dev

# Exports every variable in $(ENV_FILE) into the recipe's shell.
LOAD_ENV = test -f $(ENV_FILE) || { echo "missing $(ENV_FILE)"; exit 1; }; set -a; . ./$(ENV_FILE); set +a

help:
	@echo "Available targets:"
	@echo "  db-up           - Start only the local postgres container (.env.dev)"
	@echo "  run-server      - Run the API server on the host, loading \$$ENV_FILE (default .env.dev)"
	@echo "  run-indexer     - Run the indexer once on the host, loading \$$ENV_FILE (default .env.dev)"
	@echo "  test            - Run all Go tests"
	@echo "  lint            - Run golangci-lint"
	@echo "  build           - Build both binaries"
	@echo "  docker-up-dev   - Start postgres + API in Docker with dev env (local postgres)"
	@echo "  docker-up-prod  - Start containers with prod env (Supabase)"
	@echo "  docker-down     - Stop and remove containers"

db-up:
	docker compose --env-file .env.dev up -d --wait postgres

run-server:
	@$(LOAD_ENV); cd backend && go run ./cmd/server

run-indexer:
	@$(LOAD_ENV); cd backend && go run ./cmd/indexer

test:
	cd backend && go test -v -race ./...

lint:
	cd backend && golangci-lint run

build:
	cd backend && go build -o bin/server ./cmd/server
	cd backend && go build -o bin/indexer ./cmd/indexer

# The backend container reaches postgres by its compose service name, so override
# POSTGRES_HOST (shell variables take precedence over --env-file values).
docker-up-dev:
	POSTGRES_HOST=postgres docker compose --env-file .env.dev up -d --build

docker-up-prod:
	docker compose --env-file .env.prod up -d --build

docker-down:
	docker compose down
