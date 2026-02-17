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

	if _, err := goose.EnsureDBVersion(db); err != nil {
		return fmt.Errorf("ensure goose db version: %w", err)
	}

	beforeCount, err := appliedCount(db)
	if err != nil {
		return fmt.Errorf("get applied migrations count before: %w", err)
	}

	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}

	afterVersion, err := goose.GetDBVersion(db)
	if err != nil {
		return fmt.Errorf("get db version after migrations: %w", err)
	}

	afterCount, err := appliedCount(db)
	if err != nil {
		return fmt.Errorf("get applied migrations count after: %w", err)
	}

	appliedThisRun := afterCount - beforeCount
	if appliedThisRun < 0 {
		appliedThisRun = 0
	}

	log.Printf("db_migrations: version=%d applied=%d", afterVersion, appliedThisRun)
	return nil
}

func appliedCount(db *sql.DB) (int, error) {
	const q = `SELECT COUNT(*) FROM goose_db_version WHERE is_applied = TRUE`
	var n int
	if err := db.QueryRow(q).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}
