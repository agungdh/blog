package main

// @title Blog API
// @version 1.0
// @description API documentation for Blog SSR with admin management
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter token as: Bearer &lt;your-token&gt;

import (
	"context"
	"database/sql"
	"embed"
	"io/fs"
	"log"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

//go:embed templates/*
var templatesFS embed.FS

//go:embed static/*
var staticFS embed.FS

//go:embed migrations/*.sql
var migrationsFS embed.FS

func openDB(dbPath string) *sql.DB {
	sqldb, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

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
	return sqldb
}

func runMigrations(sqldb *sql.DB) {
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
}
