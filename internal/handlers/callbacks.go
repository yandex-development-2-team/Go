package handlers

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	. "github.com/yandex-development-2-team/Go/config"
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
	// Получаем и находим нужный handler в карте handlers
	buttonID := query.Data

	handler, ok := router.handlers[buttonID]
	if !ok {
		// Если handler не найден, возвращаем ошибку или логируем событие
		err := fmt.Errorf("oбработчик для идентификатора кнопки не найден")
		router.logger.Error("handler не найден для кнопки", zap.Error(err), zap.String("buttonID", buttonID))
		return err
	}

	// Вызываем метод Handle у найденного handler'а
	err = handler.Handle(congig.Config, query)
	if err != nil {
		return err
	}
	_, err = bot.AnswerCallbackQuery(tgbotapi.CallbackQueryID{CallbackQueryID: query.Id, Text: ""})
	return err

}
