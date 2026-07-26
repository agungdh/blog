package main

import (
	"context"
	"database/sql"
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/pressly/goose/v3"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	_ "modernc.org/sqlite"

	"blog/internal/handler"
	"blog/internal/model"
	"blog/internal/service"
	"blog/internal/store"
)

//go:embed templates/*
var templatesFS embed.FS

//go:embed static/*
var staticFS embed.FS

//go:embed migrations/*.sql
var migrationsFS embed.FS

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "blog.db"
	}

	sqldb, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer func() { _ = sqldb.Close() }()

	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA busy_timeout=5000",
		"PRAGMA synchronous=NORMAL",
		"PRAGMA cache_size=-2000",
		"PRAGMA foreign_keys=ON",
	}
	for _, p := range pragmas {
		if _, err := sqldb.Exec(p); err != nil {
			log.Fatalf("failed to apply pragma: %v", err)
		}
	}

	migrationsDir, err := fs.Sub(migrationsFS, "migrations")
	if err != nil {
		log.Fatalf("failed to create migrations sub-filesystem: %v", err)
	}
	provider, err := goose.NewProvider(goose.DialectSQLite3, sqldb, migrationsDir)
	if err != nil {
		log.Fatalf("failed to create migration provider: %v", err)
	}
	if _, err := provider.Up(context.Background()); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	db := bun.NewDB(sqldb, sqlitedialect.New())
	defer func() { _ = db.Close() }()

	st := store.New(db)

	ctx := context.Background()
	seedData(ctx, st)

	svc := service.New(st)

	tmpl, err := template.ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		log.Fatalf("failed to parse templates: %v", err)
	}

	ssr := handler.NewSSR(svc, tmpl)

	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-Forwarded-Proto") == "https" {
				r.URL.Scheme = "https"
			}
			next.ServeHTTP(w, r)
		})
	})
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.ClientIPFromRemoteAddr)

	staticSub, err := fs.Sub(staticFS, "static")
	if err != nil {
		log.Fatalf("failed to create static sub-filesystem: %v", err)
	}
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(staticSub))))

	r.Get("/", ssr.Home)
	r.Get("/posts/{slug}", ssr.Post)
	r.Get("/categories/{slug}", ssr.Category)
	r.Get("/tags/{slug}", ssr.Tag)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func seedData(ctx context.Context, st *store.Store) {
	posts, err := st.GetAllPosts(ctx)
	if err != nil {
		log.Printf("seed: failed to check existing posts: %v", err)
		return
	}
	if len(posts) > 0 {
		return
	}

	categories := []model.Category{
		{Name: "Technology", Slug: "technology"},
		{Name: "Go", Slug: "go"},
	}
	for i := range categories {
		if err := st.CreateCategory(ctx, &categories[i]); err != nil {
			log.Printf("seed: failed to create category %q: %v", categories[i].Slug, err)
		}
	}

	tags := []model.Tag{
		{Name: "go", Slug: "go"},
		{Name: "ssr", Slug: "ssr"},
		{Name: "web", Slug: "web"},
		{Name: "programming", Slug: "programming"},
	}
	for i := range tags {
		if err := st.CreateTag(ctx, &tags[i]); err != nil {
			log.Printf("seed: failed to create tag %q: %v", tags[i].Slug, err)
		}
	}

	samplePosts := []model.Post{
		{
			Slug:       "hello-world",
			Title:      "Hello World",
			Date:       time.Now().Format("2006-01-02"),
			Markdown:   "Welcome to my blog! This is the first post.\n\nThis blog is built with **Go**, using:\n\n- `net/http` and `chi` for routing\n- `html/template` for SSR\n- `goldmark` for Markdown parsing\n- `SQLite` for storage\n\nStay tuned for more posts!",
			CategoryID: categories[0].ID,
			Tags:       []model.Tag{tags[0], tags[1], tags[2]},
		},
		{
			Slug:       "getting-started",
			Title:      "Getting Started with Go SSR",
			Date:       time.Now().Add(-24 * time.Hour).Format("2006-01-02"),
			Markdown:   "## Why Go for SSR?\n\nGo is perfect for server-side rendering because:\n\n1. **Fast** compile times\n2. **Lightweight** — single binary, low memory\n3. **Built-in** `html/template` package\n4. **Efficient** — goroutines handle concurrent requests easily\n\n```go\nfunc main() {\n    http.ListenAndServe(\":8080\", r)\n}\n```\n\n> \"Simplicity is the ultimate sophistication.\" — Leonardo da Vinci",
			CategoryID: categories[1].ID,
			Tags:       []model.Tag{tags[0], tags[1], tags[3]},
		},
	}

	for i := range samplePosts {
		if err := st.CreatePost(ctx, &samplePosts[i]); err != nil {
			log.Printf("seed: failed to create post %q: %v", samplePosts[i].Slug, err)
		}
	}
	log.Printf("seeded %d posts, %d categories, %d tags", len(samplePosts), len(categories), len(tags))
}
