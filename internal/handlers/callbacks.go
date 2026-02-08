package handlers

import (
	"context"
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

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
	// Получаем и находим нужный handler в карте handlers
	button := query.Data

	handler, ok := router.handlers[button]
	if !ok {
		// Если handler не найден, возвращаем ошибку или логируем событие
		err := fmt.Errorf("oбработчик для идентификатора кнопки не найден")
		router.logger.Error("handler не найден для кнопки", zap.Error(err), zap.String("button", button))
		return err
	}

	// Вызываем метод Handle у найденного handler'а

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)s
	defer cancel()
	err := handler.Handle(ctx, query)
	if err != nil {
		return err
	}
	_, err = bot.AnswerCallbackQuery(tgbotapi.CallbackQueryID{CallbackQueryID: query.ID, Text: "Вы нажали " + button})
	end := time.Now()
	elapsed := end.Sub(start)
	router.logger.Info("Нажата кнопка "+button, zap.String("user_id", query.ID), zap.String("callback_data", button), zap.Duration("время обработки", elapsed))
	return err

}
