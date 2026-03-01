package repository

import (
	"context"
	"database/sql"
)

// DBAdapter адаптирует *sql.DB под интерфейс репозитория
type DBAdapter struct {
	db *sql.DB
}

func NewDBAdapter(db *sql.DB) *DBAdapter {
	return &DBAdapter{db: db}
}

func (a *DBAdapter) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return a.db.QueryRowContext(ctx, query, args...).Scan(dest)
}

func (a *DBAdapter) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return a.db.ExecContext(ctx, query, args...)
}

func (a *DBAdapter) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return a.db.QueryRowContext(ctx, query, args...).Scan(dest)
}

// Compile-time проверка соответствия интерфейсу
var _ DatabaseInterface = (*DBAdapter)(nil)
