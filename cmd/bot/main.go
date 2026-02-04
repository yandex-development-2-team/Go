package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"

	dbm "github.com/yandex-development-2-team/Go/internal/database"
)

func main() {
	log.Println("Bot starting...")

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Println("DATABASE_URL not set; skipping migrations")
		return
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	if err := dbm.RunMigrations(db); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}

	log.Println("Migrations applied successfully")
}
