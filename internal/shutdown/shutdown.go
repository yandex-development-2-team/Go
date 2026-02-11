package shutdown

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

type ShutdownHandler struct {
	signals chan os.Signal
	timeout time.Duration
	logger  *zap.Logger
}

type ShutdownTask struct {
	Name string
	Fn   func(context.Context) error
}

func NewShutdownHandler(logger *zap.Logger) *ShutdownHandler {
	return &ShutdownHandler{
		signals: make(chan os.Signal, 1),
		timeout: 30 * time.Second,
		logger:  logger,
	}
}

func (s *ShutdownHandler) WaitForShutdown(ctx context.Context, cancel context.CancelFunc, tasks ...ShutdownTask) error {
	signal.Notify(s.signals, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(s.signals)

	select {
	case sig := <-s.signals:
		s.logger.Info("Shutdown signal received", zap.String("signal", sig.String()))
	case <-ctx.Done():
		s.logger.Info("Context canceled before shutdown signal")
		return nil
	}

	s.logger.Info("Starting graceful shutdown")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), s.timeout)
	defer shutdownCancel()

	errCh := make(chan error, len(tasks))
	started := 0

	for _, task := range tasks {
		if task.Fn == nil {
			continue
		}

		started++

		go func(t ShutdownTask) {
			s.logger.Info("Shutdown task started", zap.String("task", t.Name))
			err := t.Fn(shutdownCtx)
			if err != nil {
				s.logger.Error("Shutdown task failed", zap.String("task", t.Name), zap.Error(err))
			} else {
				s.logger.Info("Shutdown task completed", zap.String("task", t.Name))
			}
			errCh <- err
		}(task)
	}

	var errs []error

	for i := 0; i < started; i++ {
		select {
		case err := <-errCh:
			if err != nil {
				errs = append(errs, err)
			}
		case <-shutdownCtx.Done():
			s.logger.Error("Graceful shutdown timeout exceeded", zap.Duration("timeout", s.timeout))
			errs = append(errs, errors.New("graceful shutdown timeout exceeded"))

		CollectErrors:
			for {
				select {
				case err := <-errCh:
					if err != nil {
						errs = append(errs, err)
					}
				default:
					break CollectErrors
				}
			}

			return errors.Join(errs...)
		}
	}

	if len(errs) > 0 {
		s.logger.Error("Graceful shutdown finished with errors", zap.Errors("errors", errs))
		return errors.Join(errs...)
	}

	s.logger.Info("Graceful shutdown completed successfully")
	return nil
}
