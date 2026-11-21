# Architecture & Deployment

How gopackage-chronicle is built and shipped today. Read this first when picking the project back up.
Keep it current: update this file in the same PR when a component, a host or a deploy step changes.

_Last updated: 2026-11-15_

## 1. The system at a glance

```
                        GitHub Actions (hourly)
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
| Frontend | `frontend/` | Cloudflare Worker (static assets) | `.github/workflows/frontend-deploy.yml` (after Frontend CI passes) | Push to `main` touching frontend files + manual |
| API server | `backend/cmd/server` | Fly.io app `gopackage-chronicle` | `.github/workflows/fly-deploy.yml` (after Backend CI passes) | Push to `main` touching backend files + manual |
| Indexer | `backend/cmd/indexer` | GitHub Actions runner | `.github/workflows/indexer.yml` | Hourly cron + manual |
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

- **How it ships:** `.github/workflows/frontend-deploy.yml` runs on a push to `main` that touches
  `frontend/**`, `frontend-ci.yml` or `frontend-deploy.yml`.
  1. Job `checks` re-runs Frontend CI (`frontend-check`) on the merged commit.
  2. Job `deploy` runs only if it passes: `bun run build` with `PUBLIC_API_BASE_URL` set, then
     `bunx wrangler deploy`, then a `curl` of the site.

  Cloudflare lists these deploys as "Manually deployed", because only Workers Builds deploys carry git
  info. To make them traceable, the workflow sets the deployment **message** to the PR title (or the commit
  subject for a non-merge commit, cut to 100 characters) and the version **tag** to the short commit SHA.

  Only one deploy runs at a time, and a newer push waits rather than cancelling one in progress. To
  redeploy without a code change: Actions → **Deploy frontend to Cloudflare** → **Run workflow**.
  Deploy history appears under the repo's **production-frontend** environment.
- **Build variable:** `PUBLIC_API_BASE_URL` is set in `frontend-deploy.yml`, not in Cloudflare. It is baked
  into the static bundle at build time, so changing it means redeploying. `frontend/vite.config.ts` fails
  the build when it is missing, instead of shipping a site that calls `localhost:8080`.
- **Deploy credentials (set once, GitHub repo → Settings → Secrets → Actions):**
  - `CLOUDFLARE_API_TOKEN`: create it in Cloudflare → My Profile → API Tokens from the
    **Edit Cloudflare Workers** template, scoped to this account only.
  - `CLOUDFLARE_ACCOUNT_ID`: shown on the Workers & Pages overview page.
- **PR previews:** `.github/workflows/frontend-preview.yml` runs on PRs that touch `frontend/**`. It builds,
  runs `bunx wrangler preview --name pr-<number> --json`, smoke-tests the Preview URL and keeps one PR comment
  up to date with the URL. When the PR closes, `.github/workflows/frontend-preview-cleanup.yml` deletes the
  Preview. It is a separate workflow so open PRs don't show a skipped cleanup check. Both share one
  concurrency group per PR, so closing a PR cancels a preview run in progress. Fork PRs are skipped because
  they get no secrets. Both use the same `CLOUDFLARE_*` secrets as the deploy.
- **Workers Builds: disconnected.** Previews used to come from Workers Builds, but its PR comments stopped
  appearing (the check run still passed), and it duplicated the build. Keep it disconnected (dashboard →
  Worker → Settings → Build → **Disconnect**) so GitHub Actions is the only thing that builds and deploys.
- **Manual deploy (fallback):** from `frontend/`, with `wrangler` logged in:
  `PUBLIC_API_BASE_URL=https://gopackage-chronicle.fly.dev bun run build && bunx wrangler deploy`.
- **Config in the repo:** `frontend/wrangler.jsonc`. The Worker name must match the dashboard. It serves
  `build/`, and unknown paths fall back to `index.html` so deep links like `/packages/github.com/foo/bar` work.
- **Rollback:** in the Cloudflare dashboard, open the Worker's deployments and roll back to an earlier one,
  or run `bunx wrangler rollback` from `frontend/`.

### 3.2 API server → Fly.io

- **How it ships:** `.github/workflows/fly-deploy.yml` runs on a push to `main` that touches `backend/**`,
  `infra/Dockerfile.backend`, `fly.toml`, `.dockerignore` or either backend workflow file.
  1. Job `checks` re-runs Backend CI (`backend-test`, `backend-lint`, `backend-docker`) on the merged commit.
  2. Job `deploy` runs only if all of them pass: `flyctl deploy --remote-only`, then a `curl` of `/health`.

  Only one deploy runs at a time, and a newer push waits rather than cancelling a rollout in progress. To
  redeploy without a code change: Actions → **Deploy API to Fly.io** → **Run workflow** (checks still run first).
  Deploy history appears under the repo's **production** environment.
- **Deploy token (set once):** create it with `fly tokens create deploy -a gopackage-chronicle`, then save it
  as the GitHub Actions secret `FLY_API_TOKEN`. It can only deploy this app.
- **Manual deploy (fallback):** from the **repo root**, with `flyctl` logged in, run `fly deploy`.
- **What a deploy does:** Fly reads the root `fly.toml` and builds `infra/Dockerfile.backend` (two stages:
  Go build, then a small Alpine image holding only the `server` binary). It rolls out the new machine and
  waits for the `/health` check. Migrations run when the server boots.
- **Config in the repo:** `fly.toml` at the repo root (the only Fly config). App `gopackage-chronicle`,
  region `lhr` (London, closest to Supabase in Ireland), shared-cpu-1x, 256MB, scales to zero (`min_machines_running = 0`), so the
  first request after idle is slow (cold start). `.dockerignore` keeps `frontend/` and `.env*` out of the
  build context.
- **Secrets (set once, kept by Fly):**
  ```bash
  fly secrets set \
    POSTGRES_HOST=... POSTGRES_PORT=... POSTGRES_USER=... POSTGRES_DB=... \
    'POSTGRES_PASSWORD=...' \
    ALLOWED_ORIGINS='https://gopackage-chronicle.anphse.workers.dev,https://*-gopackage-chronicle.anphse.workers.dev'
  ```
  Single-quote the password: Supabase passwords can contain `#`, `$` or `!`, which shells and `.env`
  parsers can cut off or expand. `POSTGRES_SSLMODE` defaults to `require`, which is right for Supabase.
  `ALLOWED_ORIGINS` must list the frontend's origins, otherwise the browser blocks requests (CORS). The
  second entry matches Worker Preview URLs (`pr-<number>-gopackage-chronicle.anphse.workers.dev`)
  without allowing other Workers on the account. Add any custom domain here too.
- **Check it:** `curl https://gopackage-chronicle.fly.dev/health` → `OK`. Logs: `fly logs`.
  If the server can't reach the DB it logs `failed to connect to database` at ERROR level and exits, and
  Fly restarts it up to 10 times. Requests then time out instead of returning an error.
- **Rollback:** `fly releases`, then redeploy an earlier image with `fly deploy --image <image-ref>`.
  A rollback does not undo migrations.

### 3.3 Indexer → GitHub Actions

- **How it ships:** nothing to deploy. `.github/workflows/indexer.yml` checks out `main`, builds the indexer
  and runs it every hour at minute 17 (`17 * * * *`), away from the busy top of the hour when GitHub most
  often delays or drops scheduled runs. Run it by hand from the Actions tab (**Run workflow**).
- **Secrets (GitHub repo → Settings → Secrets → Actions):** `POSTGRES_HOST`, `POSTGRES_PORT`,
  `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`.
- **Check it:** the latest run in the Actions tab. `/api/stats` on the API reports the last sync time.

### 3.4 Database → Supabase

- **Hosting:** Supabase Postgres, reached through its connection pooler. The DSN uses
  `default_query_exec_mode=simple_protocol` because the pooler needs it. Don't remove it.
- **Schema changes:** add a new `NNN_description.sql` in goose format and merge. It is applied the next time
  the API starts, which means **after the deploy workflow finishes**. Never edit an existing migration.

## 4. Checks before merging to `main`

Both workflows run on every pull request, with no path filter, so they can be required status checks.

| Workflow | Jobs | What it proves |
|----------|------|----------------|
| `.github/workflows/backend-ci.yml` | `backend-test` | `go vet`, `go build ./...`, `go test -race ./...` pass |
| | `backend-lint` | `golangci-lint` v2 passes (`backend/.golangci.yml`) |
| | `backend-docker` | The Fly.io image builds from `infra/Dockerfile.backend` |
| `.github/workflows/frontend-ci.yml` | `frontend-check` | `bun run check`, `bun run build`, `wrangler deploy --dry-run` and the Playwright E2E tests (`bun run test:e2e`, API mocked) pass |

To enforce them: GitHub → Settings → Branches → rule for `main` → **Require status checks to pass** →
add `backend-test`, `backend-lint`, `backend-docker`, `frontend-check`.

On pull requests these checks don't deploy anything. After a merge to `main`, each part deploys only when
its own files changed:

| Merge touches | Frontend deploy (`frontend-deploy.yml`) | API deploy (`fly-deploy.yml`) |
|---------------|----------------------------------|-------------------------------|
| `frontend/**` or frontend workflows | yes, after Frontend CI passes | no |
| `backend/**` (API or indexer), `infra/Dockerfile.backend`, `fly.toml`, `.dockerignore`, backend workflows | no | yes, after Backend CI passes |
| Docs or other files only | no | no |

The indexer has no deploy: its hourly run builds it from whatever is on `main`.

## 5. Environments & config

| Where | Database | Config source |
|-------|----------|---------------|
| Local, everything in Docker | docker-compose `postgres` | `.env.dev` (gitignored), `make docker-up-dev` |
| Local, Go on the host | docker-compose `postgres` | `.env.dev` (gitignored), `make db-up` + `make run-server` |
| Local backend → prod DB | Supabase | `.env.prod` (gitignored), `make docker-up-prod` |
| Production API | Supabase | Fly secrets |
| Production indexer | Supabase | GitHub Actions secrets |
| Production frontend | n/a | `PUBLIC_API_BASE_URL` in `frontend-deploy.yml` |

Local setup steps are in the [README](../README.md#getting-started).

## 6. Resuming work: quick checklist

1. `git pull`, then open the Actions tab: did the last indexer runs succeed?
2. `curl https://gopackage-chronicle.fly.dev/health`. Is the API up?
3. Open the frontend's Worker URL. Does it load data?
4. Start locally (README "Getting Started"), run `make test` and `cd frontend && bun run check`.
5. Did the last **Deploy API to Fly.io** run succeed? If not, fix it and re-run it from the Actions tab.

## 7. Known gaps

- The root `go.mod` is a leftover. The real module is `backend/go.mod`.
- The frontend is client-side only (no SSR), because it ships as a static SPA.
