# Mamikos Watcher — Plan

## Context
New, empty Go project (only README.md). Goal: a long-running Go service that periodically fetches one or more Mamikos property pages, detects room availability from page text, and writes run logs on every run. For now the run log is the only output; a real notification (Telegram, on full → available) is a future item. Config is env-only (no DB). Runs locally via Docker Compose; GitHub Actions builds an image, pushes to Docker Hub, and deploys to a VM over SSH.

## Availability rules (from user examples)
| Page text | Status | Rooms |
|---|---|---|
| `Kamar penuh` | Full | 0 |
| `Tersisa N kamar` (regex `(?i)tersisa\s+(\d+)\s+kamar`) | Available | N |
| `Banyak pilihan kamar untukmu` | Available | many (-1) |
| none matched | Unknown | — (logged as warning) |

Check order: penuh → tersisa → banyak → unknown. Match case-insensitively on extracted visible text (strip HTML tags / collapse whitespace).

## Config (env)
- `MAMIKOS_URLS` — comma-separated property URLs (required)
- `FETCH_INTERVAL` — Go duration, e.g. `15m` (default `30m`)
- `HTTP_TIMEOUT` — default `20s`
- `USER_AGENT` — the `User-Agent` HTTP header sent with each request (identifies the client). Defaults to a normal Chrome browser string, since sites often block or serve different HTML to Go's default `Go-http-client/1.1`
- `LOG_LEVEL` — `info` default
- `.env.example` committed; `.env` gitignored

## Project layout
```
cmd/watcher/main.go          # load config, start scheduler, graceful shutdown (SIGINT/SIGTERM)
internal/config/config.go    # env parsing + validation
internal/fetcher/fetcher.go  # Fetcher interface; HTTP impl with timeout + User-Agent header, returns body text
internal/parser/parser.go    # Parse(html) -> Result{Title, Status, Rooms, StatusText}
internal/parser/parser_test.go  # table tests: 3 example phrases + unknown, title extraction + fallback
internal/watcher/watcher.go  # RunOnce: fetch + parse each URL, write run logs
Dockerfile                   # multi-stage, distroless/alpine final
docker-compose.yml           # local run, env_file: .env
deploy/docker-compose.yml    # VM compose, image from Docker Hub
.github/workflows/ci.yml     # test + build
.github/workflows/deploy.yml # push to Docker Hub + SSH deploy (on main)
TODO.md                      # actionable checklist + future plans
```
Scheduler: stdlib `time.Ticker` (run once on start, then every interval) — simplest for a duration-based period; no cron lib needed. Logging: `log/slog` JSON.

### Run logs (emitted every run, regardless of status)
- `check` — **one line per property URL**: the result for that single page (url, status, rooms, duration, error if any)
- `run_complete` — **one line per run**, after all URLs are checked: a report listing each property as a **title → rent status** pair, plus total duration. Lets you scan the whole run at a glance.

Rent status: the exact text as shown on the page (the matched phrase, e.g. `Kamar penuh`, `Tersisa 2 kamar`, `Banyak pilihan kamar untukmu`), not rephrased. If no phrase matches, `unknown`; if the fetch fails, `error`.

Property title: parsed from the page (`<h1>`, falling back to `og:title` / `<title>`); if it can't be found, use the URL.

Example run with 3 URLs → 3 `check` lines + 1 `run_complete`:
```
{"time":"…10:00:00","level":"INFO","msg":"check","url":"…/kost-a","title":"Kost Melati Tipe A Tebet","status":"full","rooms":0,"duration_ms":812}
{"time":"…10:00:01","level":"INFO","msg":"check","url":"…/kost-b","title":"Kost Anggrek Setiabudi","status":"available","rooms":2,"duration_ms":640}
{"time":"…10:00:02","level":"INFO","msg":"check","url":"…/kost-c","title":"Kost Mawar Kuningan","status":"available","rooms":-1,"duration_ms":701}
{"time":"…10:00:02","level":"INFO","msg":"run_complete","duration_ms":2153,"report":[
  {"title":"Kost Melati Tipe A Tebet","status":"Kamar penuh"},
  {"title":"Kost Anggrek Setiabudi","status":"Tersisa 2 kamar"},
  {"title":"Kost Mawar Kuningan","status":"Banyak pilihan kamar untukmu"}]}
```
(Shown wrapped for readability; actual output is one JSON line.)

## Risk to verify first
Mamikos may render availability client-side (Vue). Step 1 is to `curl` a real property URL and confirm the phrases appear in raw HTML. If not, fall back to a headless browser (`chromedp`) in the fetcher — parser/watcher stay unchanged behind the `Fetcher` interface.

## TODO.md content (actionable steps)
1. [ ] Verify raw HTML of a sample property contains the availability phrases (curl)
2. [ ] `go mod init`, layout, `.gitignore`, `.env.example`
3. [ ] Config package + validation
4. [ ] Parser + table tests (Kamar penuh / Tersisa 2 kamar / Banyak pilihan kamar untukmu / unknown) and property title extraction
5. [ ] Fetcher (timeout, UA, error handling, non-200 handling)
6. [ ] Watcher: check each URL, write `check` + `run_complete` logs
7. [ ] main: ticker loop, run-on-start, graceful shutdown
8. [ ] Dockerfile + local docker-compose.yml; verify `docker compose up`
9. [ ] CI workflow: `go vet`, `go test ./...`, build
10. [ ] Deploy workflow: build & push to Docker Hub (tags `latest` + git SHA), SSH to VM, write `.env` from secrets, `docker compose pull && up -d`
11. [ ] README: setup, env vars, required GitHub secrets

**Future (not now):**
- [ ] Telegram integration: notify when a property goes full → available (needs per-URL last-state tracking + Notifier interface)
- [ ] Postgres: persist check history / last state
- [ ] Prometheus metrics endpoint + Grafana dashboards; log shipping (Loki)

## GitHub secrets needed
`DOCKERHUB_USERNAME`, `DOCKERHUB_TOKEN`, `VM_HOST`, `VM_USER`, `VM_SSH_KEY`, `MAMIKOS_URLS`, `FETCH_INTERVAL`. VM prerequisite: Docker + compose plugin installed; deploy dir e.g. `~/mamikos-watcher`.

## Verification
- `go test ./...` — parser table tests pass for all example phrases
- Watcher unit test with a fake fetcher → asserts `run_complete` report has the right title/status pairs
- `cp .env.example .env`, set a real URL + `FETCH_INTERVAL=1m`, `docker compose up --build` → see `check` + `run_complete` logs each minute
- Deploy: push to main, confirm Actions run green, `ssh vm docker compose logs -f` shows runs
