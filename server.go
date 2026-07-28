package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"path"
	"regexp"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"

	"blog/internal/config"
	"blog/internal/handler/admin"
	"blog/internal/handler/auth"
	"blog/internal/handler/blog"
	mw "blog/internal/middleware"
	"blog/internal/router"
	"blog/internal/service"
	"blog/internal/store"
)

var assetHashRe = regexp.MustCompile(`^(.+)\.[a-f0-9]{8}\.(css|js)$`)

func fileHash(fsys fs.FS, name string) string {
	b, err := fs.ReadFile(fsys, name)
	if err != nil {
		log.Fatalf("failed to read %s: %v", name, err)
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])[:8]
}

func computeAssetHashes() blog.AssetHashes {
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		log.Fatalf("failed to create static sub-filesystem: %v", err)
	}
	return blog.AssetHashes{
		CSS:   fileHash(sub, "css/style.css"),
		AppJS: fileHash(sub, "js/app.js"),
	}
}

type strippedFS struct {
	fs http.FileSystem
}

func (s strippedFS) Open(name string) (http.File, error) {
	m := assetHashRe.FindStringSubmatch(path.Base(name))
	if m != nil {
		name = path.Join(path.Dir(name), m[1]+"."+m[2])
	}
	return s.fs.Open(name)
}

func cmdServe(cfg config.Config) {
	sqldb := openDB(cfg.DBPath)
	defer func() { _ = sqldb.Close() }()

	db := bun.NewDB(sqldb, sqlitedialect.New())
	defer func() { _ = db.Close() }()

	st := store.New(db)

	svc := service.New(st)

	authSvc := service.NewAuthService(st)

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

	hashes := computeAssetHashes()

	staticSub, err := fs.Sub(staticFS, "static")
	if err != nil {
		log.Fatalf("failed to create static sub-filesystem: %v", err)
	}
	staticFileServer := http.FileServer(strippedFS{fs: http.FS(staticSub)})
	staticHandler := mw.CacheControl(http.StripPrefix("/static/", staticFileServer))

	r := router.New(router.Deps{
		SSR:    blog.NewSSR(svc, tmpl, hashes),
		Auth:   auth.NewHandler(authSvc),
		Admin:  admin.NewHandler(st),
		Static: staticHandler,
	})

	if cfg.AI.Enabled {
		startGenerator(st, cfg.AI)
	} else {
		log.Printf("post generator: disabled (set AI_ENABLED=true to enable)")
	}

	log.Printf("Server running on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}

func startGenerator(st *store.Store, aiCfg config.AIConfig) {
	if aiCfg.BaseURL == "" || aiCfg.APIKey == "" || aiCfg.Model == "" {
		log.Printf("generator: AI_* env vars not set, generator disabled")
		return
	}

	interval, err := time.ParseDuration(aiCfg.Interval)
	if err != nil {
		interval = 1 * time.Hour
	}

	aiClient := service.NewAIClient(aiCfg.BaseURL, aiCfg.APIKey, aiCfg.Model)
	gen := service.NewPostGenerator(aiClient, st)

	log.Printf("generator: enabled, interval=%v, model=%s", interval, aiCfg.Model)

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
