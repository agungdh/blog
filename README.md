# Blog

Lightweight SSR blog powered by Go, SQLite, and Chi router. Single binary, zero dependencies at runtime.

## Quick Start

```bash
git clone https://github.com/agungdh/blog.git
cd blog
go build -o blog .

./blog migrate   # run database migrations
./blog seed      # seed sample data (admin user + posts)
./blog           # start server on :8080
```

Open <http://localhost:8080>.

## Commands

| Command | Description |
|---------|-------------|
| `./blog` | Start web server |
| `./blog migrate` | Run database migrations |
| `./blog seed` | Seed all data (admin + sample posts) |
| `./blog seed user` | Seed admin user only |
| `./blog seed post` | Seed sample posts only |

## Default Admin

| Field | Value |
|-------|-------|
| Username | `admin` |
| Password | `admin123` |

Login via `POST /api/admin/login` or use the [Swagger UI](http://localhost:8080/swagger/index.html).

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP listen port |
| `DB_PATH` | `blog.db` | SQLite database path |
| `AI_ENABLED` | `false` | Enable AI post generator |
| `AI_BASE_URL` | — | OpenAI-compatible API base URL |
| `AI_API_KEY` | — | API key |
| `AI_MODEL` | — | Model name |
| `AI_INTERVAL` | `1h` | Generation interval |

## API

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/api/posts` | No | Paginated posts with filter/search |
| `GET` | `/api/categories` | No | Search categories |
| `GET` | `/api/tags` | No | Search tags |
| `POST` | `/api/admin/login` | No | Login, returns bearer token |
| `GET` | `/api/admin/me` | Bearer | Current user info |
| `DELETE` | `/api/admin/logout` | Bearer | Invalidate session |

Full API docs at `/swagger/index.html`.

## Tech Stack

- **Go** + `chi` router — SSR with `html/template`
- **SQLite** via `modernc.org/sqlite` (pure Go, no CGO)
- **Bun ORM** — database queries and migrations
- **goldmark** — markdown rendering
- **goose** — database migrations
- **bcrypt** — password hashing
