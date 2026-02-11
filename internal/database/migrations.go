package database

import (
<<<<<<< HEAD
	"database/sql"
	"fmt"
	"log"

	"github.com/pressly/goose/v3"
)

// RunMigrations применяет миграции из каталога migrations/
// Выводит текущую версию базы данных и количество примененных миграций
func RunMigrations(db *sql.DB) error {
	goose.SetDialect("postgres")

	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}

	// После применения считать текущую версию и количество примененных изменений из таблицы goose
	var version sql.NullInt64
	var count int

	if err := db.QueryRow("SELECT COALESCE(MAX(version_id),0) FROM goose_db_version WHERE is_applied = TRUE").Scan(&version); err != nil {
		return fmt.Errorf("query migration version: %w", err)
	}
	if err := db.QueryRow("SELECT COUNT(*) FROM goose_db_version WHERE is_applied = TRUE").Scan(&count); err != nil {
		return fmt.Errorf("query migration count: %w", err)
	}

	ver := int64(0)
	if version.Valid {
		ver = version.Int64
	}
	log.Printf("DB migration version: %d, applied migrations: %d", ver, count)
	return nil
=======
	"fmt"
)

func RunMigrations() error {
	return fmt.Errorf("Произошла ошибка в функции RunMigrations (БД)")
>>>>>>> 96e68a5df650fadd3caec3fbafc18e13bbc9fc93
}
