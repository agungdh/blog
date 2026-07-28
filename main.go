package main

import (
	"log"
	"os"

	"blog/internal/version"
)

func main() {
	log.Printf("blog %s", version.String())

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "blog.db"
	}

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "migrate":
			cmdMigrate(dbPath)
		case "seed":
			cmdSeed(dbPath, os.Args[2:]...)
		default:
			log.Fatalf("unknown command: %s (use: migrate, seed)", os.Args[1])
		}
		return
	}

	cmdServe(dbPath)
}
