# AGENTS.md

## Commands

```bash
./blog               # run web server (default)
./blog migrate       # run database migrations
./blog seed          # seed all (admin user + sample posts)
./blog seed user     # seed admin user only
./blog seed post     # seed sample posts only
```

## Build & Run

```bash
go build -o blog .          # single binary, everything embedded
PORT=8080 DB_PATH=blog.db ./blog
```

- Set `PORT` env (default `8080`), `DB_PATH` env (default `blog.db`)
- Migrations and seeding are separate CLI commands — no longer auto-run on server start
- AI-powered post generator: set `AI_ENABLED=true` and configure `AI_BASE_URL`, `AI_API_KEY`, `AI_MODEL`, `AI_INTERVAL` (e.g. `1h`, `30m`)
- No code generation

## Architecture

```
main.go                     → entry point, command dispatch
app.go                      → embed directives, openDB, runMigrations, swagger annotations
migrate.go                  → migrate command
seed.go                     → seed commands (cmdSeed, seedAdminUser, seedSamplePosts)
server.go                   → web server command (routes, middleware, AI generator)
docs/docs.go                → generated swagger docs (swag init)

internal/
  model/
    post.go                 → Post struct (Bun-annotated)
    category.go             → Category struct
    tag.go                  → Tag struct
    post_tag.go             → PostTag join struct
    user.go                 → User struct
    session.go              → Session struct (opaque token sessions)
  store/
    sqlite.go               → Store struct + all DB methods (Bun ORM over SQLite)
  service/
    post.go                 → PostService (markdown rendering + summaries)
    auth.go                 → AuthService (bcrypt hash, token generation, login/logout)
    aiclient.go             → AI HTTP client (OpenAI-compatible API)
    generator.go            → AI post generator
  handler/
    ssr.go                  → SSR page handlers + public API endpoints
    auth.go                 → Auth handlers (login/logout/me) + auth middleware
  version/
    version.go              → version info
```

- Module name: `blog` (short, not a full URL path)
- All resources embedded at build time via `//go:embed`:
  - `templates/*` → parsed as `html/template`
  - `static/*` → served at `/static/`
  - `migrations/*.sql` → run via `./blog migrate`
- SQLite uses these PRAGMAs on every start: WAL, busy_timeout=5000, synchronous=NORMAL, cache_size=-2000, foreign_keys=ON
- Seed data auto-populates fresh databases (idempotent: skips if data already exists)

## Routes

| Path | Handler | Auth |
|------|---------|------|
| `/` | Home — all posts | No |
| `/posts/{slug}` | Single post | No |
| `/categories/{slug}` | Redirect to `/?category=` | No |
| `/tags/{slug}` | Redirect to `/?tags=` | No |
| `/static/*` | Embedded static files | No |
| `/swagger/*` | Swagger UI | No |
| `/api/posts` | JSON posts (paginated) | No |
| `/api/categories` | JSON category search | No |
| `/api/tags` | JSON tag search | No |
| `/api/admin/login` | Login | No |
| `/api/admin/me` | Current user info | Yes (Bearer) |
| `/api/admin/logout` | Invalidate session | Yes (Bearer) |

## Auth

- Opaque bearer tokens stored in `sessions` table (multi-device support)
- Passwords hashed with bcrypt
- Default admin: `admin` / `admin123` (seeded via `./blog seed user`)
- Swagger docs available at `/swagger/index.html`
- Generate/regen swagger docs: `swag init -g app.go -d ./`

## Migrations

- Stored in `migrations/*.sql` with goose `-- +goose Up` / `-- +goose Down` annotations
- Run via `./blog migrate` command
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

Types used: `feat`, `fix`, `chore`, `ci`, `refactor`
