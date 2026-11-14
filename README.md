# gopackage-chronicle

A hobby project that tracks and displays the release history of every Go module. An indexer pulls new module versions from [`index.golang.org`](https://index.golang.org) into PostgreSQL. A small JSON API serves them, and a SvelteKit frontend displays them.

## Architecture

```
index.golang.org ──► Indexer (one-off Go binary) ──► PostgreSQL ◄── API server (Go) ◄── Frontend (SvelteKit)
```

| Component | Location | Description |
|-----------|----------|-------------|
| **Indexer** | `backend/cmd/indexer` | Fetches new module versions using a cursor-based sync and writes them to PostgreSQL. Runs once and exits. |
| **API server** | `backend/cmd/server` | Standard-library `net/http` JSON API. Runs database migrations at startup. |
| **Frontend** | `frontend/` | SvelteKit + TypeScript app (Svelte 5) that browses packages and their release timelines. |

## Tech Stack

- **Backend**: Go 1.27, `net/http` (no framework), `pgx` v5 via `database/sql`, `pressly/goose` for migrations, `log/slog`
- **Frontend**: SvelteKit, TypeScript, Bun, `adapter-static` (client-side SPA)
- **Database**: PostgreSQL 16 (Supabase in production, Docker locally)
- **Infra**: Fly.io (API), Cloudflare Workers static assets (frontend, Workers Builds), GitHub Actions (indexer), docker-compose (local dev)

## Project Structure

```
backend/                 # Go module: github.com/anphx/gopackage-chronicles
  cmd/server/            # API server entrypoint
  cmd/indexer/           # Indexer entrypoint
  internal/
    api/                 # Routes, middleware, handlers/
    database/            # Env config and connection setup
    indexer/             # index.golang.org client and sync logic
    migrations/          # Embedded goose SQL migrations
    model/               # Package and Release types
    repository/          # Data access (packages, releases, sync cursor)
frontend/
  src/
    lib/api/client.ts    # Typed API client (all fetch calls live here)
    lib/components/      # PackageCard, ReleaseTimeline, ReleaseGraph, SearchBar, skeletons
    routes/              # Home page and /packages/[name]
  wrangler.jsonc         # Cloudflare Worker static assets config
infra/
  Dockerfile.backend
  fly.toml
.github/workflows/
  indexer.yml            # Runs the indexer against the production DB
  frontend-ci.yml        # PR check: type-check, build, Wrangler dry run
```

> Note: the `go.mod` at the repo root is a leftover. The real Go module is in `backend/`, so run Go commands from there or use the `make` targets.

## Getting Started

### Prerequisites

- Go 1.27+
- [Bun](https://bun.sh) (the frontend uses Bun only, not npm, yarn or pnpm)
- Docker with Docker Compose
- Optional: [`golangci-lint`](https://golangci-lint.run) **v2** for `make lint` (v1 fails to type-check this module)

### 1. Configure environment

The backend reads its configuration from the environment. docker-compose and the Makefile docker targets load it from `.env.dev` (local) or `.env.prod` (Supabase). Both files are gitignored, so create `.env.dev` in the repo root:

```bash
POSTGRES_HOST=postgres        # use "postgres" inside docker-compose, "localhost" when running Go directly
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=gopackage_chronicle
POSTGRES_SSLMODE=disable      # the local Postgres container has no TLS
PORT=8080
```

| Variable | Default | Used by |
|----------|---------|---------|
| `POSTGRES_HOST` | `localhost` | server, indexer |
| `POSTGRES_PORT` | `5432` | server, indexer |
| `POSTGRES_USER` | `postgres` | server, indexer |
| `POSTGRES_PASSWORD` | `postgres` | server, indexer |
| `POSTGRES_DB` | `gopackage_chronicle` | server, indexer |
| `POSTGRES_SSLMODE` | `require` | server, indexer |
| `PORT` | `8080` | server |

There is no `DATABASE_URL`. The DSN is built in code from these variables.

### 2. Start the database and API

Pick one of the two options. The only difference is where the API server runs, and that decides the value of `POSTGRES_HOST`.

**Option A: everything in Docker**

```bash
make docker-up-dev     # postgres + API server at http://localhost:8080
make docker-down       # stop everything (data persists in the postgres_data volume)
```

The backend container reaches Postgres through the compose service name, so `POSTGRES_HOST` must be `postgres`. If your `.env.dev` says `localhost`, the backend container exits with `dial tcp 127.0.0.1:5432: connection refused`. Override it without editing the file:

```bash
POSTGRES_HOST=postgres make docker-up-dev
```

**Option B: Postgres in Docker, Go on the host** (faster edit/run loop)

```bash
docker compose --env-file .env.dev up -d postgres
set -a; source .env.dev; set +a
export POSTGRES_HOST=localhost
make run-server        # API on :8080, applies migrations on startup
```

`make run-server` and `make run-indexer` don't load `.env.dev` themselves. Without the exported variables they fall back to the defaults in the table above, so a `POSTGRES_DB` that differs from `gopackage_chronicle` makes them connect to the wrong database.

Check it's up:

```bash
curl http://localhost:8080/health          # OK
```

### 3. Load some data

The indexer is a one-off job, not a service. Each run syncs up to 40,000 entries (20 batches × 2000), which takes about 1.5 minutes. Run it from a shell with the same variables exported as in Option B (`POSTGRES_HOST=localhost`):

```bash
make run-indexer
```

Start the API server at least once beforehand so the migrations have created the tables. Run it again to continue from the stored cursor.

### 4. Run the frontend

```bash
cd frontend
bun install
echo "PUBLIC_API_BASE_URL=http://localhost:8080" > .env
bun run dev            # http://localhost:5173
```

## Development Commands

| Command | Description |
|---------|-------------|
| `make test` | Run all Go tests with the race detector |
| `make lint` | Run `golangci-lint` (config in `backend/.golangci.yml`) |
| `make build` | Build `backend/bin/server` and `backend/bin/indexer` |
| `cd backend && go test ./... -run TestName` | Run a single Go test |
| `cd frontend && bun run check` | Type-check the frontend (`svelte-check`) |
| `cd frontend && bun test` | Run frontend tests |
| `cd frontend && bun run build` | Production build of the frontend |

## API

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/releases` | Recent releases, including each package's path |
| `GET` | `/api/packages` | Indexed packages |
| `GET` | `/api/packages/{name...}` | Package detail with its release history. `name` is the full module path, slashes included. |
| `GET` | `/health` | Health check |

List endpoints accept `?limit=` (default 50, max 200) and `?offset=`.

```bash
curl "http://localhost:8080/api/releases?limit=5"
curl "http://localhost:8080/api/packages/github.com/spf13/cobra"
```

Middleware, from outermost to innermost: CORS → request logging → per-IP rate limit (10 req/s, burst 20) → 30s request timeout.

## Database Schema

Migrations live in `backend/internal/migrations/`. They are embedded in the server binary and applied with goose at startup.

| Table | Description |
|-------|-------------|
| `packages` | One row per unique module path |
| `releases` | One row per module version (`UNIQUE(package_id, version)`), FK to `packages` |
| `sync_cursor` | Single row holding the last synced timestamp |

Indexes: `releases(package_id)` and `releases(released_at DESC)`.

Migrations are append-only. To change the schema, add a new `NNN_description.sql` file in goose format and leave the existing files untouched.

## Indexer Behaviour

1. Reads the cursor from `sync_cursor`. On the first run (empty cursor) it starts from **2026-01-01 UTC**.
2. Fetches `index.golang.org/index?since=<cursor>&limit=2000`, up to **20 batches per run**.
3. For each batch, upserts the packages and batch-inserts the releases with `ON CONFLICT DO NOTHING`, so runs are idempotent.
4. Advances the cursor after each successful batch.
5. Stops early when a batch returns fewer than 2000 entries, which means it has caught up.

## Deployment

- **API server**: Fly.io (`shared-cpu-1x`, 256MB, scales to zero), built from `infra/Dockerfile.backend` and configured in `infra/fly.toml`.
- **Frontend**: static SPA served as a Cloudflare Worker with static assets (`frontend/wrangler.jsonc`), built and deployed by Workers Builds on every push to `main`. Build settings: root directory `frontend`, build command `bun install --frozen-lockfile && bun run build`, deploy command `bunx wrangler deploy`, and build variable `PUBLIC_API_BASE_URL` (baked in at build time). Set `ALLOWED_ORIGINS` on the API to the Worker's domain(s).
  - PRs run `.github/workflows/frontend-ci.yml` (type-check, build, Wrangler dry run); deploys happen only through Workers Builds.
- **Database**: Supabase (PostgreSQL). Use `make docker-up-prod` to run the backend container locally against it with `.env.prod`.
- **Indexer**: the GitHub Actions workflow `.github/workflows/indexer.yml` reads the `POSTGRES_*` values from repository secrets. It runs daily at 20:00 UTC and can also be triggered manually through `workflow_dispatch`.
