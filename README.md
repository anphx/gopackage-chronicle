# gopackage-chronicle

A hobby project that tracks and displays historical releases of all Go packages. It periodically indexes new module releases from `index.golang.org` and exposes them through a REST API and a SvelteKit frontend.

## Architecture

Three components live in this monorepo:

| Component | Description |
|-----------|-------------|
| **Indexer** | Go binary that fetches new releases from `index.golang.org` using a cursor-based sync strategy, storing them in PostgreSQL. Run as a one-off job. |
| **API Server** | Lightweight Go HTTP server exposing a JSON REST API. |
| **Frontend** | SvelteKit app that displays package release history. |

## Tech Stack

- **Backend** — Go 1.26, `pgx` v5, `goose` (migrations)
- **Frontend** — SvelteKit, TypeScript, Bun
- **Database** — PostgreSQL 16
- **Infra** — Docker / docker-compose, Fly.io (API), Cloudflare Pages (frontend), Supabase (hosted DB)

## Project Structure

```
backend/
  cmd/server/     # HTTP API server
  cmd/indexer/    # Release indexer
  internal/
    api/          # Routes and handlers
    database/     # DB connection and config
    indexer/      # index.golang.org client and sync logic
    migrations/   # Numbered SQL migration files (goose)
    model/        # Package and Release types
    repository/   # Data access layer (packages, releases, cursor)
frontend/
  src/
    lib/api/      # Typed API client
    components/   # Svelte components (PackageCard, ReleaseTimeline, SearchBar)
    routes/       # SvelteKit pages
infra/
  Dockerfile.backend
  Dockerfile.frontend
  fly.toml
```

## Database Schema

Three tables managed by `goose` migrations in `backend/internal/migrations/`:

| Table | Description |
|-------|-------------|
| `packages` | One row per unique module path |
| `releases` | One row per module version, FK to `packages` |
| `sync_cursor` | Single row tracking the last synced timestamp |

Indexes on `releases(package_id)` and `releases(released_at DESC)` for efficient querying.

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/releases` | List recent releases (paginated) |
| `GET` | `/api/packages` | List all indexed packages (paginated) |
| `GET` | `/api/packages/{name...}` | Package detail with release history |
| `GET` | `/health` | Health check |

All list endpoints accept `?limit=` (default 50, max 200) and `?offset=` query params.

## Indexer Behaviour

- On first run, syncs releases from 1 year ago to present
- Subsequent runs pick up from the last stored cursor timestamp
- Fetches in batches of 2000 entries, up to 10 batches per run
- Idempotent — duplicate releases are silently ignored via `ON CONFLICT DO NOTHING`

## Local Development

**Prerequisites:** Docker, Go 1.26+, Bun

### Start with Docker Compose

```bash
docker compose up --build
```

Starts two services:

- `postgres` — PostgreSQL 16 with a healthcheck
- `backend` — API server (waits for postgres healthy), runs migrations on startup, available at `http://localhost:8080`

The indexer is not a long-running service — run it manually when needed:

```bash
make run-indexer
```

### Run without Docker

```bash
# API server
make run-server

# Indexer (one-off)
make run-indexer
```

### Frontend

```bash
cd frontend
bun install
bun run dev
```

## Make Targets

```
make run-server    # Run the API server
make run-indexer   # Run the indexer once manually
make test          # Run all Go tests (verbose, with race detector)
make lint          # Run golangci-lint
make build         # Build both binaries to backend/bin/
make migrate-up    # Apply pending migrations
make migrate-down  # Roll back the last migration
```

## Deployment (WIP)

- **API Server** — deployed on Fly.io (`shared-cpu-1x`, 256MB RAM) via `infra/Dockerfile.backend`
- **Frontend** — deployed on Cloudflare Pages via `adapter-cloudflare`
- **Indexer** — intended to run as a GitHub Actions scheduled workflow (twice a week); not yet configured
- **Database** — Supabase free tier (PostgreSQL)
