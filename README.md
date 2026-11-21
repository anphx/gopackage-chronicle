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
backend/                 # Go module: github.com/anphx/gopackage-chronicle
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
    lib/components/      # Page sections and charts (ActivityChart, ModuleList, ReleaseTimeline, Pagination, ...)
    lib/*.ts             # semver, format, pagination and stats helpers (unit tested with bun test)
    routes/              # Home page and /packages/[name]
  wrangler.jsonc         # Cloudflare Worker static assets config
infra/
  Dockerfile.backend     # API image (used by Fly.io, docker-compose and Backend CI)
fly.toml                 # Fly.io app config for the API (deploy with `fly deploy` from the repo root)
.github/workflows/
  indexer.yml            # Runs the indexer against the production DB
  frontend-ci.yml        # PR check: type-check, build, Wrangler dry run
  backend-ci.yml         # PR check: vet, build, race tests, golangci-lint, Docker image build
  fly-deploy.yml         # On push to main: re-run backend CI, then deploy the API to Fly.io
docs/
  ARCHITECTURE.md        # How each part is built and deployed (start here when resuming work)
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

The backend container reaches Postgres through the compose service name, so `make docker-up-dev` sets `POSTGRES_HOST=postgres` itself, whatever `.env.dev` says.

**Option B: Postgres in Docker, Go on the host** (faster edit/run loop)

```bash
make db-up             # postgres container only, waits until healthy
make run-server        # API on :8080, applies migrations on startup
```

`make run-server` and `make run-indexer` load `.env.dev` (keep `POSTGRES_HOST=localhost` and `POSTGRES_SSLMODE=disable` in it). Use `ENV_FILE=.env.prod make run-server` to point them at another env file.

Check it's up:

```bash
curl http://localhost:8080/health          # OK
```

### 3. Load some data

The indexer is a one-off job, not a service. Each run syncs up to 40,000 entries (20 batches × 2000), which takes about 1.5 minutes. It loads `.env.dev` like `make run-server`:

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
| `cd frontend && bunx --bun playwright install chromium` | Install the browser for E2E tests (one-time setup) |
| `cd frontend && bun run test:e2e` | Run browser-based frontend E2E tests |
| `cd frontend && bun run build` | Production build of the frontend |

## API

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/releases` | Recent releases, including each package's path |
| `GET` | `/api/packages` | Indexed packages |
| `GET` | `/api/packages/{name...}` | Package detail with its release history. `name` is the full module path, slashes included. |
| `GET` | `/api/stats` | Index totals, last sync time and releases per UTC week split by kind (major, minor, patch, pre-release). `?weeks=` (default 30, max 104). |
| `GET` | `/api/stats/packages` | Modules released within the last `?weeks=` weeks (default 12, max 52) with latest version, all-time release count, recent versions and weekly counts. Filters: `?q=` (path substring), `?kind=` (`major`, `minor`, `patch`, `prerelease`), `?sort=` (`recent`, `active`, `name`). Returns `total` and per-kind counts for paging. |
| `GET` | `/health` | Health check |

List endpoints accept `?limit=` (default 50, max 200) and `?offset=`.

The kind of a release is read from the shape of its version: anything with a pre-release part (pseudo-versions included) is a pre-release, `vX.0.0` is a major, `vX.Y.0` a minor and anything else a patch.

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

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for step-by-step deploy, rollback and secrets for each part.

- **API server**: Fly.io (`shared-cpu-1x`, 256MB, scales to zero), built from `infra/Dockerfile.backend` and configured in the root `fly.toml`. Deployed by `.github/workflows/fly-deploy.yml` on pushes to `main` that touch backend files, only after Backend CI passes on the merged commit (needs the `FLY_API_TOKEN` secret). `fly deploy` from the repo root still works as a manual fallback.
  - PRs run `.github/workflows/backend-ci.yml` (vet, build, race tests, lint, Docker image build).
- **Frontend**: static SPA served as a Cloudflare Worker with static assets (`frontend/wrangler.jsonc`), built and deployed by Workers Builds on every push to `main`. Build settings: root directory `frontend`, build command `bun install --frozen-lockfile && bun run build`, deploy command `bunx wrangler deploy`, and build variable `PUBLIC_API_BASE_URL` (baked in at build time). Set `ALLOWED_ORIGINS` on the API to the Worker's domain(s).
  - PRs run `.github/workflows/frontend-ci.yml` (type-check, build, Wrangler dry run); deploys happen only through Workers Builds.
- **Database**: Supabase (PostgreSQL). Use `make docker-up-prod` to run the backend container locally against it with `.env.prod`.
- **Indexer**: the GitHub Actions workflow `.github/workflows/indexer.yml` reads the `POSTGRES_*` values from repository secrets. It runs daily at 20:00 UTC and can also be triggered manually through `workflow_dispatch`.
