package handlers

import (
	"context"
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yandex-development-2-team/Go/internal/metrics"

	"go.uber.org/zap"
)

type CallbackRouter struct {
	handlers map[string]CallbackHandler
	logger   *zap.Logger
}

type CallbackHandler interface {
	Handle(ctx context.Context, query *tgbotapi.CallbackQuery) error
}

func HandleCallback(router *CallbackRouter, query *tgbotapi.CallbackQuery) error {
	if router == nil || router.logger == nil || router.handlers == nil || query == nil {
		return fmt.Errorf("invalid callback router/query")
	}
	if metrics.Default == nil {
		return fmt.Errorf("metrics not initialized")
	}
	if query.Data == "" {
		return fmt.Errorf("empty callback data")
	}
	button := query.Data

	handler, ok := router.handlers[button]
	if !ok || handler == nil {
		return fmt.Errorf("handler not found")
	}

	start := time.Now()

	metrics.Default.ActiveUsers.Inc()
	defer metrics.Default.ActiveUsers.Dec()

	metrics.Default.CallbacksReceived.Inc()

	var handlerErr error
	defer func() {
		dur := time.Since(start).Seconds()
		metrics.Default.CallbacksProcessingDuration.Observe(dur)

		router.logger.Info("handle_callback_metrics",
			zap.String("button", button),
			zap.Float64("duration_seconds", dur),
			zap.Bool("success", handlerErr == nil),
		)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := handler.Handle(ctx, query)
	if err != nil {
		handlerErr = err
		metrics.Default.MessagesErrorsTotal.Inc()
		return err
	}
	/* //когда будет bot
	_, err = bot.AnswerCallbackQuery(tgbotapi.CallbackQueryID{CallbackQueryID: query.ID, Text: "Вы нажали " + button})
	*/
	router.logger.Info("callback handled",
		zap.String("button", button),
		zap.String("callback_id", query.ID),
	)
	return nil
}
