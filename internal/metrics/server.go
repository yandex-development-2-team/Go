package metrics

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/yandex-development-2-team/Go/internal/api"
	"go.uber.org/zap"
)

type Server struct {
	srv    *http.Server
	logger *zap.Logger
}

func NewServer(port int, db *sqlx.DB, telegram api.TelegramChecker, logger *zap.Logger) *Server {
	if logger == nil {
		logger = zap.NewNop()
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/health", api.NewHealthHandler(db, telegram, logger))

	mux.Handle("/metrics", promhttp.Handler())

	s := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return &Server{srv: s, logger: logger}
}

func (s *Server) Start() error {
	s.logger.Info("http_metrics_server_started", zap.String("addr", s.srv.Addr))
	err := s.srv.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
