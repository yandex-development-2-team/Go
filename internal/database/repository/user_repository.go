package repository

import (
	"context"
	"database/sql"
	"errors"

	"go.uber.org/zap"

	"github.com/jmoiron/sqlx"

	"github.com/yandex-development-2-team/Go/internal/models"
)

type UserRepository struct {
	db     *sqlx.DB
	logger *zap.Logger
}

func (u *UserRepository) CreateUser(ctx context.Context, telegramID int64, username, firstName, lastName string) (*models.User, error) {
	if err := ctx.Err(); err != nil {
		u.logger.Error("context cancelled before query")
		return nil, err
	}
	// select user по tg id
	var user models.User
	err := u.db.GetContext(ctx, &user, "SELECT * FROM users WHERE telegram_id = $1", telegramID)
	// если существует вернуть
	if err == nil {
		u.logger.Info("user found", zap.Int64("telegram_id", telegramID))
		return &user, nil
	}
	if err != sql.ErrNoRows {
		u.logger.Error("query error", zap.Error(err))
		return nil, err
	}
	// если не существует создать запись
	_, err = u.db.ExecContext(ctx, "INSERT INTO users (telegram_id, username, first_name, last_name) VALUES ($1, $2, $3, $4)", telegramID, username, firstName, lastName)
	if err != nil {
		u.logger.Error("error creating a user")
		return nil, err
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
	err := u.db.GetContext(ctx, &user, "SELECT * FROM users WHERE telegram_id = $1", telegramID)
	if err != nil {
		if ctx.Err() != nil {
			u.logger.Error("context cancelled")
			return nil, err
		}
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
	res, err := u.db.ExecContext(ctx, "UPDATE users SET grade = $1 WHERE telegram_id = $2", grade, telegramID)
	if err != nil {
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
	err := u.db.GetContext(ctx, &user, "SELECT * FROM users WHERE telegram_id = $1", telegramID)
	if err != nil {
		if err == sql.ErrNoRows {
			u.logger.Error("no user found")
			return false, err
		}
		u.logger.Error("query error", zap.Error(err))
		return false, err
	}
	return user.IsAdmin, nil
}
