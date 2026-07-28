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

### Public

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/api/posts` | No | Paginated posts with filter/search |
| `GET` | `/api/categories` | No | Search categories |
| `GET` | `/api/tags` | No | Search tags |

### Admin (Bearer auth required)

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/admin/login` | Login, returns bearer token |
| `GET` | `/api/admin/me` | Current user info |
| `DELETE` | `/api/admin/logout` | Invalidate session |
| `GET` | `/api/admin/categories` | List categories (cursor) |
| `POST` | `/api/admin/categories` | Create category |
| `GET` | `/api/admin/categories/{id}` | Get category |
| `PUT` | `/api/admin/categories/{id}` | Update category |
| `DELETE` | `/api/admin/categories/{id}` | Delete category |
| `GET` | `/api/admin/tags` | List tags (cursor) |
| `POST` | `/api/admin/tags` | Create tag |
| `GET` | `/api/admin/tags/{id}` | Get tag |
| `PUT` | `/api/admin/tags/{id}` | Update tag |
| `DELETE` | `/api/admin/tags/{id}` | Delete tag |
| `GET` | `/api/admin/posts` | List posts (cursor) |
| `POST` | `/api/admin/posts` | Create post |
| `GET` | `/api/admin/posts/{id}` | Get post |
| `PUT` | `/api/admin/posts/{id}` | Update post |
| `DELETE` | `/api/admin/posts/{id}` | Delete post |

Full API docs at `/swagger/index.html`.

## Tech Stack

- **Go** + `chi` router — SSR with `html/template`
- **SQLite** via `modernc.org/sqlite` (pure Go, no CGO)
- **Bun ORM** — database queries, migrations, and transactions
- **go-playground/validator** — form validation
- **goldmark** — markdown rendering
- **goose** — database migrations
- **bcrypt** — password hashing
