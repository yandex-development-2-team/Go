package shutdown

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type ShutdownHandler struct {
	signals chan os.Signal
	timeout time.Duration
}

type ShutdownTask struct {
	Name string
	Fn   func(context.Context) error
}

func NewShutdownHandler() *ShutdownHandler {
	return &ShutdownHandler{
		signals: make(chan os.Signal, 1),
		timeout: 30 * time.Second,
	}
}

func (s *ShutdownHandler) WaitForShutdown(ctx context.Context, cancel context.CancelFunc, tasks ...ShutdownTask) error {
	signal.Notify(s.signals, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(s.signals)

	select {
	case <-s.signals:
		//logger
	case <-ctx.Done():
		//logger
		return nil
	}

	//logger
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
			//logger
			err := t.Fn(shutdownCtx)
			if err != nil {
				//logger
			} else {
				//logger
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
			//logger
			errs = append(errs, errors.New("graceful shutdown тайм-аут превышен"))

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
		//logger
		return errors.Join(errs...)
	}

	//logger
	return nil
}
