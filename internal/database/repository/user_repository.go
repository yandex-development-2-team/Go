package repository

import (
	"go.uber.org/zap"

	"github.com/jmoiron/sqlx"
)

type UserRepository struct {
	db     *sqlx.DB
	logger *zap.Logger
}
