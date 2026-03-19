package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/yandex-development-2-team/Go/internal/models"
)

var ErrEmailExists = errors.New("email already exists")

type AuthUserRepository struct {
	db *sqlx.DB
}

func NewAuthUserRepository(db *sqlx.DB) *AuthUserRepository {
	return &AuthUserRepository{db: db}
}

func (r *AuthUserRepository) GetByEmail(ctx context.Context, email string) (*models.AuthUser, error) {
	var u models.AuthUser
	err := r.db.GetContext(ctx, &u, "SELECT * FROM auth_users WHERE email = $1", email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &u, nil
}

func (r *AuthUserRepository) Create(ctx context.Context, name, email, passwordHash, role string) (*models.AuthUser, error) {
	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var u models.AuthUser
	err = tx.GetContext(ctx, &u, `
INSERT INTO auth_users (name, email, password_hash, role, status)
VALUES ($1, $2, $3, $4, 'active')
RETURNING id, name, email, role, status, created_at, updated_at
`, name, email, passwordHash, role)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return nil, ErrEmailExists
		}
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &u, nil
}
