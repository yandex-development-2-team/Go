package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/yandex-development-2-team/Go/internal/models"
)

type ServiceRepository struct {
	db *sqlx.DB
}

func NewServiceRepository(db *sqlx.DB) *ServiceRepository {
	return &ServiceRepository{db: db}
}

func (s *ServiceRepository) GetServicesOfBoxSolutions(ctx context.Context) ([]models.Service, error) {
	var services []models.Service
	err := s.db.SelectContext(ctx, &services, `SELECT id, title FROM services ORDER BY id`)
	return services, err
}
