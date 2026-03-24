package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/yandex-development-2-team/Go/internal/models"
	"go.uber.org/zap"
)

type ResourceService interface {
	GetBySlug(ctx context.Context, slug string) (*models.ResourcePage, error)
	UpdatePartial(ctx context.Context, slug string, req models.ResourcePageUpdateRequest) (*models.ResourcePage, error)
}

type ResourcesHandler struct {
	service ResourceService
	logger  *zap.Logger
}

func NewResourcesHandler(service ResourceService, logger *zap.Logger) http.Handler {
	return &ResourcesHandler{service: service, logger: logger}
}

func (h *ResourcesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	slug := cleanSlug(r.URL.Path, "/api/v1/resources/")
	if slug == "" {
		respondValidationError(w, "slug", "slug is required")
		return
	}

	switch r.Method {
	case http.MethodPut:
		h.handleUpdate(w, r, slug)
	case http.MethodGet:
		h.handleGet(w, r, slug)
	default:
		respondError(w, http.StatusMethodNotAllowed, "method not allowed", "")
	}
}

func (h *ResourcesHandler) handleUpdate(w http.ResponseWriter, r *http.Request, slug string) {
	contentType := r.Header.Get("Content-Type")
	if !strings.Contains(strings.ToLower(contentType), "application/json") {
		respondValidationError(w, "content_type", "Content-Type must be application/json")
		return
	}

	defer r.Body.Close()

	var req models.ResourcePageUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondValidationError(w, "body", "invalid json")
		return
	}

	if req.Title == nil && req.Content == nil && req.Links == nil {
		respondValidationError(w, "body", "one field must be provided: title, content or links")
		return
	}

	if req.Links != nil {
		for i, link := range *req.Links {
			if link.Title == "" {
				respondValidationError(w, fmt.Sprintf("links[%d].title", i), "title is required")
				return
			}
			if link.URL == "" {
				respondValidationError(w, fmt.Sprintf("links[%d].url", i), "url is required")
				return
			}
			u, err := url.ParseRequestURI(link.URL)
			if err != nil || u.Scheme == "" || u.Host == "" {
				msg := "invalid URI (scheme and host are required)"
				if err != nil {
					msg = fmt.Sprintf("invalid URI: %s", err)
				}
				respondValidationError(w, fmt.Sprintf("links[%d].url", i), msg)
				return
			}
		}
	}

	updated, err := h.service.UpdatePartial(r.Context(), slug, req)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "resource not found", "")
		return
	}
	if err != nil {
		h.logger.Error("update failed", zap.Error(err), zap.String("slug", slug))
		respondError(w, http.StatusInternalServerError, "internal server error", "")
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

func (h *ResourcesHandler) handleGet(w http.ResponseWriter, r *http.Request, slug string) {
	page, err := h.service.GetBySlug(r.Context(), slug)
	if err != nil {
		h.logger.Error("get resource failed", zap.Error(err), zap.String("slug", slug))
		respondError(w, http.StatusInternalServerError, "internal error", "")
		return
	}
	if page == nil {
		respondError(w, http.StatusNotFound, "not found", "")
		return
	}

	writeJSON(w, http.StatusOK, page)
}
