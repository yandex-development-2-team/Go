package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
)

var (
	ErrRefreshNotFound = errors.New("refresh token not found")
	ErrRefreshRevoked  = errors.New("refresh token revoked")
	ErrRefreshExpired  = errors.New("refresh token expired")
)

type RefreshTokenRepository struct {
	db *sqlx.DB
}

func NewRefreshTokenRepository(db *sqlx.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

type RefreshTokenRow struct {
	Token     string     `db:"token"`
	UserID    int64      `db:"user_id"`
	ExpiresAt time.Time  `db:"expires_at"`
	RevokedAt *time.Time `db:"revoked_at"`
}

func (r *RefreshTokenRepository) Insert(ctx context.Context, token string, userID int64, expiresAt time.Time) error {
	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(ctx, `
INSERT INTO refresh_tokens (token, user_id, expires_at)
VALUES ($1, $2, $3)
`, token, userID, expiresAt)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *RefreshTokenRepository) Revoke(ctx context.Context, token string) error {
	res, err := r.db.ExecContext(ctx, `UPDATE refresh_tokens SET revoked_at = NOW() WHERE token = $1 AND revoked_at IS NULL`, token)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrRefreshRevoked
	}
	return nil
}

// Rotate: блокируем токен, валидируем, ревоким, создаём новый
func (r *RefreshTokenRepository) Rotate(ctx context.Context, token string, newToken string, newExpiresAt time.Time) (int64, error) {
	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	var row RefreshTokenRow
	err = tx.GetContext(ctx, &row, `SELECT token, user_id, expires_at, revoked_at FROM refresh_tokens WHERE token = $1 FOR UPDATE`, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrRefreshNotFound
		}
		return 0, err
	}
	if row.RevokedAt != nil {
		return 0, ErrRefreshRevoked
	}
	if time.Now().After(row.ExpiresAt) {
		return 0, ErrRefreshExpired
	}

	_, err = tx.ExecContext(ctx, `UPDATE refresh_tokens SET revoked_at = NOW() WHERE token = $1`, token)
	if err != nil {
		return 0, err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO refresh_tokens (token, user_id, expires_at) VALUES ($1, $2, $3)`, newToken, row.UserID, newExpiresAt)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return row.UserID, nil
}
