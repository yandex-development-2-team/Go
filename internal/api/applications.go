package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/yandex-development-2-team/Go/internal/models"
	"go.uber.org/zap"
)

type ApplicationsService interface {
	CreateApplication(context.Context, models.ApplicationCreateRequest) (*models.Application, error)
	ListApplications(context.Context, models.ApplicationFilter) ([]models.ApplicationListItem, int, error)
}

func NewApplicationsHandler(service ApplicationsService, logger *zap.Logger) http.HandlerFunc {
	if logger == nil {
		logger = zap.NewNop()
	}

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case http.MethodPost:
			handleApplicationsCreate(w, r, service, logger)
		case http.MethodGet:
			handleApplicationsList(w, r, service, logger)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func handleApplicationsCreate(w http.ResponseWriter, r *http.Request, service ApplicationsService, logger *zap.Logger) {
	var req models.ApplicationCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if !isValidApplicationType(req.Type) {
		http.Error(w, "invalid type", http.StatusBadRequest)
		return
	}
	if !isValidApplicationSource(req.Source) {
		http.Error(w, "invalid source", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.CustomerName) == "" {
		http.Error(w, "customer_name is required", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.ContactInfo) == "" {
		http.Error(w, "contact_info is required", http.StatusBadRequest)
		return
	}

	app, err := service.CreateApplication(r.Context(), req)
	if err != nil {
		logger.Error("failed to create application", zap.Error(err))
		http.Error(w, "failed to create application", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(app)
}

func handleApplicationsList(w http.ResponseWriter, r *http.Request, service ApplicationsService, logger *zap.Logger) {
	q := r.URL.Query()

	var filter models.ApplicationFilter

	if statusStr := q.Get("status"); statusStr != "" {
		status := models.ApplicationStatus(statusStr)
		if !isValidApplicationStatus(status) {
			http.Error(w, "invalid status", http.StatusBadRequest)
			return
		}
		filter.Status = &status
	}

	if typeStr := q.Get("type"); typeStr != "" {
		atype := models.ApplicationType(typeStr)
		if !isValidApplicationType(atype) {
			http.Error(w, "invalid type", http.StatusBadRequest)
			return
		}
		filter.Type = &atype
	}

	if managerIdStr := q.Get("manager_id"); managerIdStr != "" {
		mid, err := strconv.ParseInt(managerIdStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid manager_id", http.StatusBadRequest)
			return
		}
		filter.ManagerID = &mid
	}

	if dateFrom := q.Get("date_from"); dateFrom != "" {
		parsed, err := parseDate(dateFrom)
		if err != nil {
			http.Error(w, "invalid date_from", http.StatusBadRequest)
			return
		}
		filter.DateFrom = parsed
	}

	if dateTo := q.Get("date_to"); dateTo != "" {
		parsed, err := parseDate(dateTo)
		if err != nil {
			http.Error(w, "invalid date_to", http.StatusBadRequest)
			return
		}
		filter.DateTo = parsed
	}

	limit := 20
	if limitStr := q.Get("limit"); limitStr != "" {
		parsed, err := strconv.Atoi(limitStr)
		if err != nil || parsed <= 0 {
			http.Error(w, "invalid limit", http.StatusBadRequest)
			return
		}
		limit = parsed
	}

	offset := 0
	if offsetStr := q.Get("offset"); offsetStr != "" {
		parsed, err := strconv.Atoi(offsetStr)
		if err != nil || parsed < 0 {
			http.Error(w, "invalid offset", http.StatusBadRequest)
			return
		}
		offset = parsed
	}

	filter.Limit = limit
	filter.Offset = offset

	items, total, err := service.ListApplications(r.Context(), filter)
	if err != nil {
		logger.Error("failed to list applications", zap.Error(err))
		http.Error(w, "failed to list applications", http.StatusInternalServerError)
		return
	}

	resp := struct {
		Items      []models.ApplicationListItem `json:"items"`
		Pagination models.Pagination            `json:"pagination"`
	}{
		Items: items,
		Pagination: models.Pagination{
			Total:  total,
			Limit:  limit,
			Offset: offset,
		},
	}

	_ = json.NewEncoder(w).Encode(resp)
}

func parseDate(value string) (*time.Time, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return &t, nil
	}
	if t, err := time.Parse("2006-01-02", value); err == nil {
		// Interpret date boundary as start of day for from and end of day for to to be easy,
		// but since list query uses exact <= we keep as 23:59:59 for to if needed.
		return &t, nil
	}
	return nil, fmt.Errorf("invalid date format")
}

func isValidApplicationType(t models.ApplicationType) bool {
	return t == models.ApplicationTypeBox || t == models.ApplicationTypeSpecialProject
}

func isValidApplicationSource(s models.ApplicationSource) bool {
	return s == models.ApplicationSourceTelegramBot || s == models.ApplicationSourceManual
}

func isValidApplicationStatus(s models.ApplicationStatus) bool {
	return s == models.ApplicationStatusQueue || s == models.ApplicationStatusInProgress || s == models.ApplicationStatusDone
}
