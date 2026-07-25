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

	seedData(st)

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
	r.Get("/categories/{slug}", ssr.Category)
	r.Get("/tags/{slug}", ssr.Tag)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func seedData(st *store.Store) {
	posts, err := st.GetAllPosts()
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
		if err := st.CreateCategory(&categories[i]); err != nil {
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
		if err := st.CreateTag(&tags[i]); err != nil {
			log.Printf("seed: failed to create tag %q: %v", tags[i].Slug, err)
		}
	}

	techCategoryID := categories[0].ID
	goCategoryID := categories[1].ID

	goTag := tags[0]
	ssrTag := tags[1]
	webTag := tags[2]
	progTag := tags[3]

	samplePosts := []model.Post{
		{
			Slug:       "hello-world",
			Title:      "Hello World",
			Date:       time.Now().Format("2006-01-02"),
			Markdown:   "Welcome to my blog! This is the first post.\n\nThis blog is built with **Go**, using:\n\n- `net/http` and `chi` for routing\n- `html/template` for SSR\n- `goldmark` for Markdown parsing\n- `SQLite` for storage\n\nStay tuned for more posts!",
			CategoryID: techCategoryID,
			Tags:       []model.Tag{goTag, ssrTag, webTag},
		},
		{
			Slug:       "getting-started",
			Title:      "Getting Started with Go SSR",
			Date:       time.Now().Add(-24 * time.Hour).Format("2006-01-02"),
			Markdown:   "## Why Go for SSR?\n\nGo is perfect for server-side rendering because:\n\n1. **Fast** compile times\n2. **Lightweight** — single binary, low memory\n3. **Built-in** `html/template` package\n4. **Efficient** — goroutines handle concurrent requests easily\n\n```go\nfunc main() {\n    http.ListenAndServe(\":8080\", r)\n}\n```\n\n> \"Simplicity is the ultimate sophistication.\" — Leonardo da Vinci",
			CategoryID: goCategoryID,
			Tags:       []model.Tag{goTag, ssrTag, progTag},
		},
	}

	for i := range samplePosts {
		if err := st.CreatePost(&samplePosts[i]); err != nil {
			log.Printf("seed: failed to create post %q: %v", samplePosts[i].Slug, err)
		}
	}
	log.Printf("seeded %d posts, %d categories, %d tags", len(samplePosts), len(categories), len(tags))
}
