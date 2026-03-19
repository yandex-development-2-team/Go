package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/yandex-development-2-team/Go/internal/models"
	"go.uber.org/zap"
)

type ApplicationRepository struct {
	db     DatabaseInterface
	logger *zap.Logger
}

func NewApplicationRepository(db DatabaseInterface, logger *zap.Logger) *ApplicationRepository {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &ApplicationRepository{db: db, logger: logger}
}

func (r *ApplicationRepository) CreateApplication(ctx context.Context, req models.ApplicationCreateRequest) (*models.Application, error) {
	if err := ctx.Err(); err != nil {
		r.logger.Error("context cancelled before create application")
		return nil, err
	}

	const query = `
INSERT INTO applications (type, source, status, customer_name, contact_info, project_name, box_id, special_project_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, type, source, status, customer_name, contact_info, project_name, box_id, special_project_id, manager_id, created_at, updated_at
`

	app := &models.Application{}
	if err := r.db.GetContext(ctx, app, query,
		req.Type,
		req.Source,
		models.ApplicationStatusQueue,
		req.CustomerName,
		req.ContactInfo,
		req.ProjectName,
		req.BoxID,
		req.SpecialProjectID,
	); err != nil {
		r.logger.Error("failed to insert application", zap.Error(err))
		return nil, err
	}

	return app, nil
}

func (r *ApplicationRepository) ListApplications(ctx context.Context, filter models.ApplicationFilter) ([]models.ApplicationListItem, int, error) {
	if err := ctx.Err(); err != nil {
		r.logger.Error("context cancelled before list applications")
		return nil, 0, err
	}

	conditions := []string{"1=1"}
	args := []interface{}{}
	argIdx := 1

	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *filter.Status)
		argIdx++
	}
	if filter.Type != nil {
		conditions = append(conditions, fmt.Sprintf("type = $%d", argIdx))
		args = append(args, *filter.Type)
		argIdx++
	}
	if filter.ManagerID != nil {
		conditions = append(conditions, fmt.Sprintf("manager_id = $%d", argIdx))
		args = append(args, *filter.ManagerID)
		argIdx++
	}
	if filter.DateFrom != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIdx))
		args = append(args, *filter.DateFrom)
		argIdx++
	}
	if filter.DateTo != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIdx))
		args = append(args, *filter.DateTo)
		argIdx++
	}

	where := strings.Join(conditions, " AND ")

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM applications WHERE %s", where)
	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		r.logger.Error("failed to count applications", zap.Error(err))
		return nil, 0, err
	}

	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	pageQuery := fmt.Sprintf(`
SELECT id, type, source, status, customer_name, contact_info, project_name, box_id, special_project_id, manager_id, created_at, updated_at
FROM applications
WHERE %s
ORDER BY created_at DESC
LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)

	args = append(args, filter.Limit, filter.Offset)

	items := []models.ApplicationListItem{}
	if err := r.db.SelectContext(ctx, &items, pageQuery, args...); err != nil {
		r.logger.Error("failed to list applications", zap.Error(err))
		return nil, 0, err
	}

	return items, total, nil
}
