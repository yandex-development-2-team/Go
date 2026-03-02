package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/yandex-development-2-team/Go/internal/metrics"
	"go.uber.org/zap"

	"github.com/yandex-development-2-team/Go/internal/models"
)

const (
	dbQueryTimeout     = 10 * time.Second
	slowQueryThreshold = 1 * time.Second
)

type UserRepository struct {
	db     DatabaseInterface
	logger *zap.Logger
}

func NewUserRepository(db DatabaseInterface, logger *zap.Logger) *UserRepository {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &UserRepository{
		db:     db,
		logger: logger,
	}
}

func (u *UserRepository) CreateUser(ctx context.Context, telegramID int64, username, firstName, lastName string) (*models.User, error) {
	if err := ctx.Err(); err != nil {
		u.logger.Error("context cancelled before query")
		return nil, err
	}
	// select user по tg id
	var user models.User
	op := "read"
	ctxQ, cancel := context.WithTimeout(ctx, dbQueryTimeout)
	start := time.Now()
	err := u.db.GetContext(ctxQ, &user, "SELECT * FROM users WHERE telegram_id = $1", telegramID)
	dur := time.Since(start).Seconds()
	cancel()

	metrics.DBQueriesTotal.WithLabelValues(op).Inc()
	metrics.DBQueryDuration.WithLabelValues(op).Observe(dur)
	if dur > slowQueryThreshold.Seconds() {
		u.logger.Warn("slow_db_query", zap.String("operation", op), zap.Float64("duration_seconds", dur))
	}
	// если существует вернуть
	if err == nil {
		u.logger.Info("user found", zap.Int64("telegram_id", telegramID))
		return &user, nil
	}
	if err != sql.ErrNoRows {
		metrics.DBErrorsTotal.WithLabelValues(op).Inc()
		u.logger.Error("query error", zap.Error(err))
		return nil, err
	}
	// если не существует создать запись
	op = "create"
	ctxQ, cancel = context.WithTimeout(ctx, dbQueryTimeout)
	start = time.Now()
	var userID int64
	err = u.db.GetContext(
		ctxQ,
		&userID,
		"INSERT INTO users (telegram_id, username, first_name, last_name) VALUES ($1, $2, $3, $4) RETURNING id",
		telegramID, username, firstName, lastName,
	)
	dur = time.Since(start).Seconds()
	cancel()

	metrics.DBQueriesTotal.WithLabelValues(op).Inc()
	metrics.DBQueryDuration.WithLabelValues(op).Observe(dur)
	if dur > slowQueryThreshold.Seconds() {
		u.logger.Warn("slow_db_query", zap.String("operation", op), zap.Float64("duration_seconds", dur))
	}
	if err != nil {
		metrics.DBErrorsTotal.WithLabelValues(op).Inc()
		u.logger.Error("error creating a user", zap.Error(err))
		return nil, err
	}

	op = "read"
	ctxQ, cancel = context.WithTimeout(ctx, dbQueryTimeout)
	start = time.Now()
	err = u.db.GetContext(ctxQ, &user, "SELECT * FROM users WHERE ID = $1", userID)
	dur = time.Since(start).Seconds()
	cancel()

	metrics.DBQueriesTotal.WithLabelValues(op).Inc()
	metrics.DBQueryDuration.WithLabelValues(op).Observe(dur)
	if dur > slowQueryThreshold.Seconds() {
		u.logger.Warn("slow_db_query", zap.String("operation", op), zap.Float64("duration_seconds", dur))
	}
	if err != nil && err != sql.ErrNoRows {
		metrics.DBErrorsTotal.WithLabelValues(op).Inc()
	}

	u.logger.Info("user created", zap.Int64("telegram_id", telegramID))
	return &user, nil
}

func (u *UserRepository) GetUserByTelegramID(ctx context.Context, telegramID int64) (*models.User, error) {
	if err := ctx.Err(); err != nil {
		u.logger.Error("context cancelled before query")
		return nil, err
	}
	var user models.User
	op := "read"
	ctxQ, cancel := context.WithTimeout(ctx, dbQueryTimeout)
	start := time.Now()
	err := u.db.GetContext(ctxQ, &user, "SELECT * FROM users WHERE telegram_id = $1", telegramID)
	dur := time.Since(start).Seconds()
	cancel()

	metrics.DBQueriesTotal.WithLabelValues(op).Inc()
	metrics.DBQueryDuration.WithLabelValues(op).Observe(dur)
	if dur > slowQueryThreshold.Seconds() {
		u.logger.Warn("slow_db_query", zap.String("operation", op), zap.Float64("duration_seconds", dur))
	}

	if err != nil && err != sql.ErrNoRows {
		if ctx.Err() != nil {
			u.logger.Error("context cancelled")
			return nil, err
		}
		metrics.DBErrorsTotal.WithLabelValues(op).Inc()
		u.logger.Error("query error", zap.Error(err))
		return nil, err
	}
	u.logger.Info("user found", zap.Int64("telegram_id", telegramID))
	return &user, nil
}

func (u *UserRepository) UpdateUserGrade(ctx context.Context, telegramID int64, grade int) error {
	if err := ctx.Err(); err != nil {
		u.logger.Error("context cancelled before query")
		return err
	}
	op := "update"
	ctxQ, cancel := context.WithTimeout(ctx, dbQueryTimeout)
	start := time.Now()
	res, err := u.db.ExecContext(ctxQ, "UPDATE users SET grade = $1 WHERE telegram_id = $2", grade, telegramID)
	dur := time.Since(start).Seconds()
	cancel()

	metrics.DBQueriesTotal.WithLabelValues(op).Inc()
	metrics.DBQueryDuration.WithLabelValues(op).Observe(dur)
	if dur > slowQueryThreshold.Seconds() {
		u.logger.Warn("slow_db_query", zap.String("operation", op), zap.Float64("duration_seconds", dur))
	}
	if err != nil {
		metrics.DBErrorsTotal.WithLabelValues(op).Inc()
		u.logger.Error("query error", zap.Error(err))
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		u.logger.Error("failed to get rows affected", zap.Error(err))
		return err
	}
	if rowsAffected == 0 {
		u.logger.Error("no user found")
		return errors.New("no user found")
	}
	u.logger.Info("user grade updated succesfully")
	return nil
}

func (u *UserRepository) IsAdmin(ctx context.Context, telegramID int64) (bool, error) {
	if err := ctx.Err(); err != nil {
		u.logger.Error("context cancelled before query")
		return false, err
	}
	var user models.User
	op := "read"
	ctxQ, cancel := context.WithTimeout(ctx, dbQueryTimeout)
	start := time.Now()
	err := u.db.GetContext(ctxQ, &user, "SELECT * FROM users WHERE telegram_id = $1", telegramID)
	dur := time.Since(start).Seconds()
	cancel()

	metrics.DBQueriesTotal.WithLabelValues(op).Inc()
	metrics.DBQueryDuration.WithLabelValues(op).Observe(dur)
	if dur > slowQueryThreshold.Seconds() {
		u.logger.Warn("slow_db_query", zap.String("operation", op), zap.Float64("duration_seconds", dur))
	}

	if err != nil {
		if err == sql.ErrNoRows {
			u.logger.Error("no user found")
			return false, err
		}
		metrics.DBErrorsTotal.WithLabelValues(op).Inc()
		u.logger.Error("query error", zap.Error(err))
		return false, err
	}
	return user.IsAdmin, nil
}
