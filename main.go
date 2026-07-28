package main

import (
	"log"
	"os"

	"blog/internal/config"
	"blog/internal/version"
)

func main() {
	log.Printf("blog %s", version.String())

	cfg := config.Load()

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "migrate":
			cmdMigrate(cfg.DBPath)
		case "seed":
			cmdSeed(cfg.DBPath, os.Args[2:]...)
		default:
			log.Fatalf("unknown command: %s (use: migrate, seed)", os.Args[1])
		}
		return
	}

	cmdServe(cfg)
}
