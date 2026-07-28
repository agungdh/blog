package main

import "log"

func cmdMigrate(dbPath string) {
	sqldb := openDB(dbPath)
	defer func() { _ = sqldb.Close() }()

	runMigrations(sqldb)
	log.Println("migrations complete")
}
