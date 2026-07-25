package main

import (
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"blog/internal/handler"
	"blog/internal/model"
	"blog/internal/service"
	"blog/internal/store"
)

//go:embed templates/*
var templatesFS embed.FS

//go:embed static/*
var staticFS embed.FS

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "blog.db"
	}

	st, err := store.New(dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer st.Close()

	seedPosts(st)

	svc := service.New(st)

	tmpl, err := template.ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		log.Fatalf("failed to parse templates: %v", err)
	}

	ssr := handler.NewSSR(svc, tmpl)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)

	staticSub, err := fs.Sub(staticFS, "static")
	if err != nil {
		log.Fatalf("failed to create static sub-filesystem: %v", err)
	}
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(staticSub))))

	r.Get("/", ssr.Home)
	r.Get("/posts/{slug}", ssr.Post)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func seedPosts(st *store.Store) {
	posts, err := st.GetAllPosts()
	if err != nil {
		log.Printf("seed: failed to check existing posts: %v", err)
		return
	}
	if len(posts) > 0 {
		return
	}

	samplePosts := []model.Post{
		{
			Slug:     "hello-world",
			Title:    "Hello World",
			Date:     time.Now().Format("2006-01-02"),
			Markdown: "Welcome to my blog! This is the first post.\n\nThis blog is built with **Go**, using:\n\n- `net/http` and `chi` for routing\n- `html/template` for SSR\n- `goldmark` for Markdown parsing\n- `SQLite` for storage\n\nStay tuned for more posts!",
		},
		{
			Slug:     "getting-started",
			Title:    "Getting Started with Go SSR",
			Date:     time.Now().Add(-24 * time.Hour).Format("2006-01-02"),
			Markdown: "## Why Go for SSR?\n\nGo is perfect for server-side rendering because:\n\n1. **Fast** compile times\n2. **Lightweight** — single binary, low memory\n3. **Built-in** `html/template` package\n4. **Efficient** — goroutines handle concurrent requests easily\n\n```go\nfunc main() {\n    http.ListenAndServe(\":8080\", r)\n}\n```\n\n> \"Simplicity is the ultimate sophistication.\" — Leonardo da Vinci",
		},
	}

	for _, p := range samplePosts {
		post := p
		if err := st.CreatePost(&post); err != nil {
			log.Printf("seed: failed to create post %q: %v", p.Slug, err)
		}
	}
	log.Printf("seeded %d sample posts", len(samplePosts))
}
