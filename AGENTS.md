# AGENTS.md

## Build & Run

```bash
go build -o blog .          # single binary, everything embedded
PORT=8080 DB_PATH=blog.db ./blog
```

- Set `PORT` env (default `8080`), `DB_PATH` env (default `blog.db`)
- AI-powered post generator: set `AI_ENABLED=true` and configure `AI_BASE_URL`, `AI_API_KEY`, `AI_MODEL`, `AI_INTERVAL` (e.g. `1h`, `30m`)
- No code generation

## Architecture

```
main.go
  internal/handler/ssr.go   → chi HTTP handlers + Go templates
  internal/service/post.go  → markdown rendering (goldmark) + summaries
  internal/store/sqlite.go   → Bun ORM queries over SQLite
  internal/model/*.go        → Bun-annotated structs (Post, Category, Tag, PostTag)
```

- Module name: `blog` (short, not a full URL path)
- All resources embedded at build time via `//go:embed`:
  - `templates/*` → parsed as `html/template`
  - `static/*` → served at `/static/`
  - `migrations/*.sql` → auto-applied by goose on startup
- SQLite uses these PRAGMAs on every start: WAL, busy_timeout=5000, synchronous=NORMAL, cache_size=-2000, foreign_keys=ON
- Seed data auto-populates on fresh databases (when post count == 0)

## Routes

| Path | Handler |
|------|--------|
| `/` | Home — all posts |
| `/posts/{slug}` | Single post |
| `/categories/{slug}` | Posts by category |
| `/tags/{slug}` | Posts by tag |
| `/static/*` | Embedded static files |

## Migrations

- Stored in `migrations/*.sql` with goose `-- +goose Up` / `-- +goose Down` annotations
- Run automatically on startup via `goose.NewProvider`
- To add a migration: create a new numbered SQL file in `migrations/`

## Testing & Linting

- No `_test.go` files exist. There is currently no test suite.
- No linter or formatter config (no `.golangci.yml`, etc.)
- CI only builds — no test or lint jobs

## CI/CD

- Push to `master`/`main` or tag `v*` → cross-compiles 5 platforms (`CGO_ENABLED=0`, `-ldflags="-s -w"`)
- Tag push → auto-upload binaries to GitHub Release via `gh release upload`
- Targets: linux/amd64, linux/arm64, windows/amd64, darwin/amd64, darwin/arm64

## Commit Convention

```
type: description
```

Types used: `feat`, `fix`, `chore`, `ci`
