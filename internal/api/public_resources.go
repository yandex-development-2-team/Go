package api

import (
	"net/http"

	"go.uber.org/zap"
)

func NewPublicResourcesHandler(service ResourceService, logger *zap.Logger) http.Handler {
	h := &ResourcesHandler{service: service, logger: logger}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			respondError(w, http.StatusMethodNotAllowed, "method not allowed", "")
			return
		}

		slug := cleanSlug(r.URL.Path, "/public/resources/")
		if slug == "" {
			respondValidationError(w, "slug", "slug is required")
			return
		}

		h.handleGet(w, r, slug)
	})
}
