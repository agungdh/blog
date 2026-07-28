package main

import (
	"context"
	"log"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"

	"blog/internal/model"
	"blog/internal/service"
	"blog/internal/store"
)

func cmdSeed(dbPath string, args ...string) {
	sqldb := openDB(dbPath)
	defer sqldb.Close()

	runMigrations(sqldb)

	db := bun.NewDB(sqldb, sqlitedialect.New())
	defer db.Close()

	st := store.New(db)
	ctx := context.Background()

	what := "all"
	if len(args) > 0 {
		what = args[0]
	}

	switch what {
	case "user":
		seedAdminUser(ctx, st)
	case "post":
		seedSamplePosts(ctx, st)
	case "all":
		seedAdminUser(ctx, st)
		seedSamplePosts(ctx, st)
	default:
		log.Fatalf("unknown seed target: %s (use: user, post, all)", what)
	}

	log.Println("seed complete")
}

func seedSamplePosts(ctx context.Context, st *store.Store) {
	count, err := st.CountPosts(ctx)
	if err != nil {
		log.Printf("seed: failed to check existing posts: %v", err)
		return
	}
	if count > 0 {
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

func seedAdminUser(ctx context.Context, st *store.Store) {
	count, err := st.CountUsers(ctx)
	if err != nil {
		log.Printf("seed: failed to check existing users: %v", err)
		return
	}
	if count > 0 {
		return
	}

	authSvc := service.NewAuthService(st)
	hash, err := authSvc.HashPassword("admin123")
	if err != nil {
		log.Printf("seed: failed to hash password: %v", err)
		return
	}

	user := &model.User{
		Username: "admin",
		Password: hash,
		Nama:     "Administrator",
	}
	if err := st.CreateUser(ctx, user); err != nil {
		log.Printf("seed: failed to create admin user: %v", err)
		return
	}
	log.Printf("seeded default admin user (username: admin, password: admin123)")
}
