package main

import (
	"context"
	"crypto/tls"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"

	httpSwagger "github.com/swaggo/http-swagger"

	_ "blog/docs"

	"blog/internal/handler/admin"
	"blog/internal/handler/auth"
	"blog/internal/handler/blog"
	"blog/internal/service"
	"blog/internal/store"
)

func cmdServe(dbPath string) {
	sqldb := openDB(dbPath)
	defer func() { _ = sqldb.Close() }()

	db := bun.NewDB(sqldb, sqlitedialect.New())
	defer func() { _ = db.Close() }()

	st := store.New(db)

	svc := service.New(st)

	authSvc := service.NewAuthService(st)
	authHandler := auth.NewHandler(authSvc)
	adminHandler := admin.NewHandler(st)

	tmpl := template.New("").Funcs(template.FuncMap{
		"contains": func(slice []string, item string) bool {
			for _, s := range slice {
				if s == item {
					return true
				}
			}
			return false
		},
	})
	tmpl, err := tmpl.ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		log.Fatalf("failed to parse templates: %v", err)
	}

	ssr := blog.NewSSR(svc, tmpl)

	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			proto := r.Header.Get("X-Forwarded-Proto")
			if proto == "" {
				for _, fwd := range r.Header.Values("Forwarded") {
					for _, part := range strings.Split(fwd, ";") {
						kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
						if len(kv) == 2 && strings.TrimSpace(kv[0]) == "proto" {
							proto = strings.Trim(kv[1], `"`)
						}
					}
				}
			}
			if proto == "https" {
				r.TLS = &tls.ConnectionState{}
			}
			next.ServeHTTP(w, r)
		})
	})
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

	r.Get("/api/posts", ssr.APIPosts)
	r.Get("/api/categories", ssr.APISearchCategories)
	r.Get("/api/tags", ssr.APISearchTags)

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Route("/api/admin", func(r chi.Router) {
		r.Post("/login", authHandler.Login)

		r.Group(func(r chi.Router) {
			r.Use(authHandler.AuthMiddleware)

			r.Get("/me", authHandler.Me)
			r.Delete("/logout", authHandler.Logout)

			r.Get("/categories", adminHandler.ListCategories)
			r.Post("/categories", adminHandler.CreateCategory)
			r.Get("/categories/{id}", adminHandler.GetCategory)
			r.Put("/categories/{id}", adminHandler.UpdateCategory)
			r.Delete("/categories/{id}", adminHandler.DeleteCategory)

			r.Get("/tags", adminHandler.ListTags)
			r.Post("/tags", adminHandler.CreateTag)
			r.Get("/tags/{id}", adminHandler.GetTag)
			r.Put("/tags/{id}", adminHandler.UpdateTag)
			r.Delete("/tags/{id}", adminHandler.DeleteTag)

			r.Get("/posts", adminHandler.ListPosts)
			r.Post("/posts", adminHandler.CreatePost)
			r.Get("/posts/{id}", adminHandler.GetPost)
			r.Put("/posts/{id}", adminHandler.UpdatePost)
			r.Delete("/posts/{id}", adminHandler.DeletePost)
		})
	})

	r.NotFound(ssr.NotFound)

	if os.Getenv("AI_ENABLED") == "true" {
		startGenerator(st)
	} else {
		log.Printf("post generator: disabled (set AI_ENABLED=true to enable)")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func startGenerator(st *store.Store) {
	baseURL := os.Getenv("AI_BASE_URL")
	apiKey := os.Getenv("AI_API_KEY")
	model := os.Getenv("AI_MODEL")
	if baseURL == "" || apiKey == "" || model == "" {
		log.Printf("generator: AI_* env vars not set, generator disabled")
		return
	}

	intervalStr := os.Getenv("AI_INTERVAL")
	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		interval = 1 * time.Hour
	}

	aiClient := service.NewAIClient(baseURL, apiKey, model)
	gen := service.NewPostGenerator(aiClient, st)

	log.Printf("generator: enabled, interval=%v, model=%s", interval, model)

	go func() {
		ctx := context.Background()
		gen.Generate(ctx)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			gen.Generate(ctx)
		}
	}()
}
