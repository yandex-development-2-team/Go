package api

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// HealthStatus represents the health check response
type HealthStatus struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	DB        string    `json:"db"`
	Telegram  string    `json:"telegram"`
	Details   *Details  `json:"details,omitempty"`
}

type Details struct {
	DBLatency       string `json:"db_latency,omitempty"`
	TelegramLatency string `json:"telegram_latency,omitempty"`
	TelegramCached  bool   `json:"telegram_cached,omitempty"`
}

// TelegramChecker interface for checking Telegram API
type TelegramChecker interface {
	Ping(ctx context.Context) error
}

// HealthHandler handles health check requests
type HealthHandler struct {
	db             *sqlx.DB
	telegram       TelegramChecker
	logger         *zap.Logger
	telegramCache  *telegramCache
	includeDetails bool
}

// Cache for Telegram check results
type telegramCache struct {
	mu        sync.RWMutex
	status    string
	latency   time.Duration
	checkedAt time.Time
	ttl       time.Duration
}

func newTelegramCache(ttl time.Duration) *telegramCache {
	return &telegramCache{
		status: "unknown",
		ttl:    ttl,
	}
}

func (c *telegramCache) get() (string, time.Duration, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if time.Since(c.checkedAt) < c.ttl {
		return c.status, c.latency, true
	}
	return "", 0, false
}

func (c *telegramCache) set(status string, latency time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.status = status
	c.latency = latency
	c.checkedAt = time.Now()
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *sqlx.DB, telegram TelegramChecker, logger *zap.Logger) http.HandlerFunc {
	h := &HealthHandler{
		db:             db,
		telegram:       telegram,
		logger:         logger,
		telegramCache:  newTelegramCache(time.Minute),
		includeDetails: true,
	}
	return h.ServeHTTP
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	response := HealthStatus{
		Status:    "ok",
		Timestamp: time.Now().UTC(),
		DB:        "ok",
		Telegram:  "ok",
	}

	var details Details
	allHealthy := true

	// Check database
	dbStatus, dbLatency := h.checkDB(ctx)
	response.DB = dbStatus
	details.DBLatency = dbLatency.String()
	if dbStatus != "ok" {
		allHealthy = false
	}

	// Check Telegram (with cache)
	if h.telegram != nil {
		tgStatus, tgLatency, cached := h.checkTelegram(ctx)
		response.Telegram = tgStatus
		details.TelegramLatency = tgLatency.String()
		details.TelegramCached = cached
		if tgStatus != "ok" {
			allHealthy = false
		}
	} else {
		response.Telegram = "not_configured"
	}

	// Set overall status
	if !allHealthy {
		response.Status = "degraded"
	}

	// Include details if enabled
	if h.includeDetails {
		response.Details = &details
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	// Set status code
	statusCode := http.StatusOK
	if !allHealthy {
		statusCode = http.StatusServiceUnavailable
	}
	w.WriteHeader(statusCode)

	// Log health check
	h.logger.Info("health check",
		zap.String("status", response.Status),
		zap.String("db", response.DB),
		zap.String("telegram", response.Telegram),
		zap.Int("http_status", statusCode),
	)

	// Encode response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode health response", zap.Error(err))
	}
}

func (h *HealthHandler) checkDB(ctx context.Context) (string, time.Duration) {
	start := time.Now()

	// Simple ping query
	err := h.db.PingContext(ctx)
	latency := time.Since(start)

	if err != nil {
		h.logger.Error("DB ping failed", zap.Error(err))
		return "error", time.Since(start)
	}

	// Optional: run a simple query to verify full connectivity
	var result int
	err = h.db.GetContext(ctx, &result, "SELECT 1")
	if err != nil {
		h.logger.Error("DB SELECT 1 failed", zap.Error(err))
		return "error", time.Since(start)
	}

	h.logger.Debug("DB check passed", zap.Duration("latency", time.Since(start)))
	return "ok", latency

}

func (h *HealthHandler) checkTelegram(ctx context.Context) (string, time.Duration, bool) {
	// Check cache first
	if status, latency, cached := h.telegramCache.get(); cached {
		h.logger.Debug("Telegram check from cache", zap.String("status", status), zap.Duration("latency", latency))
		return status, latency, true
	}

	// Perform actual check
	start := time.Now()
	err := h.telegram.Ping(ctx)
	latency := time.Since(start)

	status := "ok"
	if err != nil {
		h.logger.Error("Telegram ping failed", zap.Error(err))
		status = "error"
	}
	h.telegramCache.set(status, latency)
	return status, latency, false

}
