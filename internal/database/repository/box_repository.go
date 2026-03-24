package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/yandex-development-2-team/Go/internal/metrics"
	"github.com/yandex-development-2-team/Go/internal/models"
)

type BoxRepository struct {
	db     DatabaseInterface
	logger *zap.Logger
}

func NewBoxRepository(db DatabaseInterface, logger *zap.Logger) *BoxRepository {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &BoxRepository{
		db:     db,
		logger: logger,
	}
}

var validBoxStatuses = map[string]bool{
	"active":    true,
	"hidden":    true,
	"draft":     true,
	"processed": true,
}

func ValidateStatus(status string) bool {
	return validBoxStatuses[status]
}

func (b *BoxRepository) SelectByID(ctx context.Context, id int) (*models.Box, error) {
	if err := ctx.Err(); err != nil {
		b.logger.Error("context cancelled before query")
		return nil, err
	}
	var box models.Box
	op := "read"
	ctxQ, cancel := context.WithTimeout(ctx, dbQueryTimeout)
	start := time.Now()
	err := b.db.GetContext(ctxQ, &box, "SELECT * FROM boxes WHERE id = $1 AND deleted_at IS NULL", id)
	dur := time.Since(start).Seconds()
	cancel()

	metrics.Default.DatabaseQueriesTotal.WithLabelValues(op).Inc()
	metrics.Default.DatabaseQueryDuration.WithLabelValues(op).Observe(dur)
	if dur > slowQueryThreshold.Seconds() {
		b.logger.Warn("slow_db_query", zap.String("operation", op), zap.Float64("duration_seconds", dur))
	}
	if err == nil {
		b.logger.Info("box found", zap.Int64("id", box.ID))
		return &box, nil
	}
	if err == sql.ErrNoRows {
		metrics.Default.DatabaseErrorsTotal.WithLabelValues(op).Inc()
		b.logger.Error("box not found", zap.Error(err))
		return nil, err
	}
	metrics.Default.DatabaseErrorsTotal.WithLabelValues(op).Inc()
	b.logger.Error("query error", zap.Error(err))
	return nil, err
}

func (b *BoxRepository) Update(ctx context.Context, id int64, updateData map[string]interface{}) error {
	if err := ctx.Err(); err != nil {
		b.logger.Error("context cancelled before query")
		return err
	}
	if status, exists := updateData["status"]; exists {
		statusStr, ok := status.(string)
		if !ok {
			return fmt.Errorf("status must be a string")
		}

		if !ValidateStatus(statusStr) {
			b.logger.Error("invalid status", zap.String("status", statusStr))
			return fmt.Errorf("invalid status: %s", statusStr)
		}
	}
	updateData["updated_at"] = time.Now()
	op := "update"
	ctxQ, cancel := context.WithTimeout(ctx, dbQueryTimeout)
	start := time.Now()
	setParts := make([]string, 0, len(updateData))
	args := make([]interface{}, 0, len(updateData)+1)
	argIndex := 1

	for field, value := range updateData {
		setParts = append(setParts, fmt.Sprintf("%s = $%d", field, argIndex))
		args = append(args, value)
		argIndex++
	}
	args = append(args, id)

	query := fmt.Sprintf(`
        UPDATE boxes 
        SET %s 
        WHERE id = $%d AND deleted_at IS NULL
    `, strings.Join(setParts, ", "), argIndex)

	_, err := b.db.ExecContext(ctxQ, query, args...)
	dur := time.Since(start).Seconds()
	cancel()

	metrics.Default.DatabaseQueriesTotal.WithLabelValues(op).Inc()
	metrics.Default.DatabaseQueryDuration.WithLabelValues(op).Observe(dur)

	if err != nil {
		metrics.Default.DatabaseErrorsTotal.WithLabelValues(op).Inc()
		b.logger.Error("query error", zap.Error(err))
		return err
	}
	return nil
}

func (b *BoxRepository) Delete(ctx context.Context, id int64) error {
	if err := ctx.Err(); err != nil {
		b.logger.Error("context cancelled before query")
		return err
	}
	op := "delete"
	ctxQ, cancel := context.WithTimeout(ctx, dbQueryTimeout)
	start := time.Now()
	_, err := b.db.ExecContext(ctxQ, "UPDATE boxes SET deleted_at = $1 WHERE id = $2", time.Now(), id)
	dur := time.Since(start).Seconds()
	cancel()

	metrics.Default.DatabaseQueriesTotal.WithLabelValues(op).Inc()
	metrics.Default.DatabaseQueryDuration.WithLabelValues(op).Observe(dur)

	if err != nil {
		metrics.Default.DatabaseErrorsTotal.WithLabelValues(op).Inc()
		b.logger.Error("query error", zap.Error(err))
		return err
	}

	return nil
}

func (b *BoxRepository) UpdateStatus(ctx context.Context, id int64, status string) (*models.Box, error) {
	if err := ctx.Err(); err != nil {
		b.logger.Error("context cancelled before query")
		return nil, err
	}
	var box models.Box

	op := "update"
	ctxQ, cancel := context.WithTimeout(ctx, dbQueryTimeout)
	start := time.Now()

	_, err := b.db.ExecContext(ctxQ, "UPDATE boxes SET status = $1, updated_at = $2 WHERE id = $3 AND deleted_at IS NULL RETURNING", status, time.Now(), id)
	dur := time.Since(start).Seconds()
	cancel()

	metrics.Default.DatabaseQueriesTotal.WithLabelValues(op).Inc()
	metrics.Default.DatabaseQueryDuration.WithLabelValues(op).Observe(dur)

	if err != nil {
		metrics.Default.DatabaseErrorsTotal.WithLabelValues(op).Inc()
		b.logger.Error("query error", zap.Error(err))
		return nil, err
	}

	err = b.db.GetContext(ctxQ, &box, "SELECT * FROM boxes WHERE id = $1 AND deleted_at IS NULL", id)

	if err != nil {
		if err == sql.ErrNoRows {
			metrics.Default.DatabaseErrorsTotal.WithLabelValues(op).Inc()
			b.logger.Error("box not found", zap.Error(err))
			return nil, fmt.Errorf("box with id %d not found", id)
		}
		metrics.Default.DatabaseErrorsTotal.WithLabelValues(op).Inc()
		b.logger.Error("query error", zap.Error(err))
		return nil, err
	}

	return &box, nil
}

func (b *BoxRepository) GetBoxesForExport(ctx context.Context, statusFilter *string) ([]models.Box, error) {
	if err := ctx.Err(); err != nil {
		b.logger.Error("context cancelled before query")
		return nil, err
	}

	op := "read"
	ctxQ, cancel := context.WithTimeout(ctx, dbQueryTimeout)
	start := time.Now()

	var boxes []models.Box
	var err error

	if statusFilter != nil && *statusFilter != "" {
		err = b.db.SelectContext(ctxQ, &boxes, "SELECT * FROM boxes WHERE status = $1 AND deleted_at IS NULL ORDER BY created_at DESC", *statusFilter)
	} else {
		err = b.db.SelectContext(ctxQ, &boxes, "SELECT * FROM boxes WHERE deleted_at IS NULL ORDER BY created_at DESC")
	}

	dur := time.Since(start).Seconds()
	cancel()

	metrics.Default.DatabaseQueriesTotal.WithLabelValues(op).Inc()
	metrics.Default.DatabaseQueryDuration.WithLabelValues(op).Observe(dur)

	if err != nil {
		metrics.Default.DatabaseErrorsTotal.WithLabelValues(op).Inc()
		b.logger.Error("query error", zap.Error(err))
		return nil, err
	}

	return boxes, nil
}
