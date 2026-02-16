package database

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/pressly/goose/v3"
)

const migrationsDir = "migrations"

func RunMigrations(db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}

	_, err := goose.EnsureDBVersion(db)
	if err != nil {
		return fmt.Errorf("ensure goose db version: %w", err)
	}

	beforeVersion, err := goose.GetDBVersion(db)
	if err != nil {
		return fmt.Errorf("get db version before migrations: %w", err)
	}

	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}

	afterVersion, err := goose.GetDBVersion(db)
	if err != nil {
		return fmt.Errorf("get db version after migrations: %w", err)
	}

	appliedThisRun := int(afterVersion - beforeVersion)
	if appliedThisRun < 0 {
		appliedThisRun = 0
	}

	log.Printf("db_migrations: version=%d applied=%d", afterVersion, appliedThisRun)
	return nil
}
