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
	start := time.Now()

	metrics.Default.ActiveUsers.Inc()
	defer metrics.Default.ActiveUsers.Dec()

	metrics.Default.CallbacksReceived.Inc()
	// Получаем и находим нужный handler в карте handlers
	button := query.Data

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

	handler, ok := router.handlers[button]
	if !ok {
		// Если handler не найден, возвращаем ошибку или логируем событие
		err := fmt.Errorf("oбработчик для идентификатора кнопки не найден")
		router.logger.Error("handler не найден для кнопки", zap.Error(err), zap.String("button", button))
		return err
	}

	// Вызываем метод Handle у найденного handler'а

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
	return err

}
