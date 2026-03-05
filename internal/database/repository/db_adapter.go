package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// DBAdapter адаптирует *sql.DB под интерфейс репозитория
type DBAdapter struct {
	db *sqlx.DB
}

func NewDBAdapter(db *sql.DB) *DBAdapter {
	return &DBAdapter{db: sqlx.NewDb(db, "postgres")}
}

func (a *DBAdapter) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return a.db.GetContext(ctx, dest, query, args...)
}

func (a *DBAdapter) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return a.db.ExecContext(ctx, query, args...)
}

func (a *DBAdapter) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return a.db.SelectContext(ctx, dest, query, args...)
}

// Compile-time проверка соответствия интерфейсу
var _ DatabaseInterface = (*DBAdapter)(nil)
