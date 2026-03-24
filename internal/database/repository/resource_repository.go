package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/yandex-development-2-team/Go/internal/models"
	"go.uber.org/zap"
)

type ResourceRepository struct {
	db     *sqlx.DB
	logger *zap.Logger
}

func NewResourceRepository(db *sqlx.DB, logger *zap.Logger) *ResourceRepository {
	return &ResourceRepository{db: db, logger: logger}
}

func (r *ResourceRepository) GetBySlug(ctx context.Context, slug string) (*models.ResourcePage, error) {
	var page models.ResourcePage
	err := r.db.GetContext(ctx, &page,
		`SELECT slug, title, content, links, created_at, updated_at 
         FROM resource_pages WHERE slug = $1`, slug)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		r.logger.Error("get resource by slug failed", zap.Error(err), zap.String("slug", slug))
		return nil, err
	}
	return &page, nil
}

func (r *ResourceRepository) UpdatePartial(ctx context.Context, slug string, req models.ResourcePageUpdateRequest) (*models.ResourcePage, error) {
	var page models.ResourcePage

	var linksArg interface{}
	if req.Links != nil {
		linksArg = models.ResourceLinks(*req.Links)
	} else {
		linksArg = nil
	}

	err := r.db.GetContext(ctx, &page, `
		UPDATE resource_pages
		SET 
			title      = COALESCE($1, title),
			content    = COALESCE($2, content),
			links      = COALESCE($3, links),
			updated_at = NOW()
		WHERE slug = $4
		RETURNING slug, title, content, links, created_at, updated_at`,
		req.Title,
		req.Content,
		linksArg,
		slug,
	)

	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		r.logger.Error("failed to update resource",
			zap.Error(err),
			zap.String("slug", slug))
		return nil, err
	}

	r.logger.Info("resource updated successfully",
		zap.String("slug", slug),
		zap.Bool("title_updated", req.Title != nil),
		zap.Bool("content_updated", req.Content != nil),
		zap.Bool("links_updated", req.Links != nil),
		zap.Int("links_count", len(page.Links)),
		zap.Time("updated_at", page.UpdatedAt),
	)

	return &page, nil
}
