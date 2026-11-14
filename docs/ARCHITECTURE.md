# Architecture & Deployment

How gopackage-chronicle is built and shipped today. Read this first when picking the project back up.
Keep it current: update this file in the same PR when a component, a host or a deploy step changes.

_Last updated: 2026-11-14_

## 1. The system at a glance

```
                        GitHub Actions (daily 20:00 UTC)
                                     │
index.golang.org ──► Indexer (Go, one-off) ──┐
                                             ▼
                                     PostgreSQL 16 (Supabase)
                                             ▲
Browser ──► Frontend (static SPA) ──► API server (Go)
            Cloudflare Worker          Fly.io
```

| Part | Code | Runs on | Deployed by | Trigger |
|------|------|---------|-------------|---------|
| Frontend | `frontend/` | Cloudflare Worker (static assets) | Cloudflare Workers Builds | Automatic on push to `main` |
| API server | `backend/cmd/server` | Fly.io app `gopackage-chronicle` | **You, by hand** (`fly deploy` from repo root) | Manual |
| Indexer | `backend/cmd/indexer` | GitHub Actions runner | `.github/workflows/indexer.yml` | Daily cron + manual |
| Database | `backend/internal/migrations` | Supabase (Postgres 16) | API server runs migrations at startup | On every API start |

The browser talks only to the API. Only the API and the indexer talk to the database.

## 2. Components

**Frontend.** SvelteKit + Svelte 5 + TypeScript, built with Bun. `adapter-static` makes it a fully static
single-page app. Every API call goes through `src/lib/api/client.ts`. The API URL
(`PUBLIC_API_BASE_URL`) is **baked in at build time**, so changing it means rebuilding.

**API server.** Go standard-library `net/http`, `pgx` through `database/sql`, `log/slog`. Read-only JSON
endpoints under `/api/*` plus `/health`. Applies goose migrations (embedded in the binary) on startup.
Middleware: CORS → logging → per-IP rate limit → 30s timeout.

**Indexer.** A Go binary that runs once and exits. It reads a cursor from the `sync_cursor` table, pulls up
to 20 × 2000 entries from `index.golang.org`, inserts them with `ON CONFLICT DO NOTHING` (safe to re-run)
and advances the cursor. It does **not** run migrations, so the API must have started at least once
against a new database first.

**Database.** Three tables: `packages`, `releases`, `sync_cursor`. Migrations are append-only files in
`backend/internal/migrations/`.

## 3. Deployment, part by part

### 3.1 Frontend → Cloudflare Workers

- **How it ships:** Cloudflare Workers Builds is connected to the GitHub repo. A push to `main` builds and
  deploys production. Pushes to other branches get preview builds (`wrangler preview`).
  No GitHub Action deploys the frontend.
- **Build settings (in the Cloudflare dashboard → Worker → Settings → Build, not in the repo):**
  - Root directory: `frontend`
  - Build command: `bun install --frozen-lockfile && bun run build`
  - Deploy command: `bunx wrangler deploy`
  - Build variable: `PUBLIC_API_BASE_URL` = `https://gopackage-chronicle.fly.dev`. It must be a **build**
    variable (Settings → Build → Variables and secrets), not a runtime Worker variable
    (Settings → Variables and secrets): the static site has no runtime to read it.
    `frontend/vite.config.ts` fails the build when it is missing, instead of shipping a site that
    calls `localhost:8080`.
- **Config in the repo:** `frontend/wrangler.jsonc`. The Worker name must match the dashboard. It serves
  `build/`, and unknown paths fall back to `index.html` so deep links like `/packages/github.com/foo/bar` work.
- **Rollback:** in the Cloudflare dashboard, open the Worker's deployments and roll back to an earlier one.

### 3.2 API server → Fly.io (manual)

- **How it ships:** there is no automatic deploy. From the **repo root**, with `flyctl` logged in:
  ```bash
  fly deploy
  ```
  Fly reads the root `fly.toml` and builds `infra/Dockerfile.backend` (two stages: Go build, then a small
  Alpine image holding only the `server` binary). It rolls out the new machine and waits for the `/health`
  check. Migrations run when the server boots.
- **Config in the repo:** `fly.toml` at the repo root (the only Fly config). App `gopackage-chronicle`,
  region `lhr` (London, closest to Supabase in Ireland), shared-cpu-1x, 256MB, scales to zero (`min_machines_running = 0`), so the
  first request after idle is slow (cold start). `.dockerignore` keeps `frontend/` and `.env*` out of the
  build context.
- **Secrets (set once, kept by Fly):**
  ```bash
  fly secrets set \
    POSTGRES_HOST=... POSTGRES_PORT=... POSTGRES_USER=... POSTGRES_DB=... \
    'POSTGRES_PASSWORD=...' \
    ALLOWED_ORIGINS=https://<worker-domain>
  ```
  Single-quote the password: Supabase passwords can contain `#`, `$` or `!`, which shells and `.env`
  parsers can cut off or expand. `POSTGRES_SSLMODE` defaults to `require`, which is right for Supabase.
  `ALLOWED_ORIGINS` must list the frontend's domain(s), otherwise the browser blocks requests (CORS).
- **Check it:** `curl https://gopackage-chronicle.fly.dev/health` → `OK`. Logs: `fly logs`.
  If the server can't reach the DB it logs `failed to connect to database` at ERROR level and exits, and
  Fly restarts it up to 10 times. Requests then time out instead of returning an error.
- **Rollback:** `fly releases`, then redeploy an earlier image with `fly deploy --image <image-ref>`.
  A rollback does not undo migrations.

### 3.3 Indexer → GitHub Actions

- **How it ships:** nothing to deploy. `.github/workflows/indexer.yml` checks out `main`, builds the indexer
  and runs it, daily at 20:00 UTC (`0 20 * * *`). Run it by hand from the Actions tab (**Run workflow**).
- **Secrets (GitHub repo → Settings → Secrets → Actions):** `POSTGRES_HOST`, `POSTGRES_PORT`,
  `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`.
- **Check it:** the latest run in the Actions tab. `/api/stats` on the API reports the last sync time.

### 3.4 Database → Supabase

- **Hosting:** Supabase Postgres, reached through its connection pooler. The DSN uses
  `default_query_exec_mode=simple_protocol` because the pooler needs it. Don't remove it.
- **Schema changes:** add a new `NNN_description.sql` in goose format and merge. It is applied the next time
  the API starts, which means **after you run `fly deploy`**. Never edit an existing migration.

## 4. Checks before merging to `main`

Both workflows run on every pull request, with no path filter, so they can be required status checks.

| Workflow | Jobs | What it proves |
|----------|------|----------------|
| `.github/workflows/backend-ci.yml` | `backend-test` | `go vet`, `go build ./...`, `go test -race ./...` pass |
| | `backend-lint` | `golangci-lint` v2 passes (`backend/.golangci.yml`) |
| | `backend-docker` | The Fly.io image builds from `infra/Dockerfile.backend` |
| `.github/workflows/frontend-ci.yml` | `frontend-check` | `bun run check`, `bun run build`, `wrangler deploy --dry-run` pass |

To enforce them: GitHub → Settings → Branches → rule for `main` → **Require status checks to pass** →
add `backend-test`, `backend-lint`, `backend-docker`, `frontend-check`.

These checks don't deploy anything. Merging to `main` deploys the frontend (Workers Builds) but **not**
the API, which still needs `fly deploy`.

## 5. Environments & config

| Where | Database | Config source |
|-------|----------|---------------|
| Local, everything in Docker | docker-compose `postgres` | `.env.dev` (gitignored), `make docker-up-dev` |
| Local, Go on the host | docker-compose `postgres` | export `.env.dev` in the shell, `make run-server` |
| Local backend → prod DB | Supabase | `.env.prod` (gitignored), `make docker-up-prod` |
| Production API | Supabase | Fly secrets |
| Production indexer | Supabase | GitHub Actions secrets |
| Production frontend | n/a | Workers Builds variable `PUBLIC_API_BASE_URL` |

Local setup steps are in the [README](../README.md#getting-started).

## 6. Resuming work: quick checklist

1. `git pull`, then open the Actions tab: did the last indexer runs succeed?
2. `curl https://gopackage-chronicle.fly.dev/health`. Is the API up?
3. Open the frontend's Worker URL. Does it load data?
4. Start locally (README "Getting Started"), run `make test` and `cd frontend && bun run check`.
5. Merged backend changes since the last deploy? Run `fly deploy` from the repo root.

## 7. Known gaps

- The API has no automatic deploy. A Fly deploy workflow on push to `main` (with a `FLY_API_TOKEN` secret)
  would close this.
- The root `go.mod` is a leftover. The real module is `backend/go.mod`.
- The frontend is client-side only (no SSR), because it ships as a static SPA.
