package handlers

import (
	"context"
	"fmt"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type ButtonHandlerOne struct{}

func (b *ButtonHandlerOne) Handle(ctx context.Context, q *tgbotapi.CallbackQuery) error {
	fmt.Println("Запуск хендлера1")
	return nil
}

type ButtonHandlerTwo struct{}

func (b *ButtonHandlerTwo) Handle(ctx context.Context, q *tgbotapi.CallbackQuery) error {
	fmt.Println("Запуск хендлера2")
	return nil
}
func TestHandleCallback(t *testing.T) {
	logger := zap.NewExample()
	defer logger.Sync()

	router := &CallbackRouter{
		handlers: make(map[string]CallbackHandler),
		logger:   logger,
	}

	router.handlers["button1"] = &ButtonHandlerOne{}
	router.handlers["button2"] = &ButtonHandlerTwo{}

	query := &tgbotapi.CallbackQuery{
		ID:   "your_callback_query_id",
		Data: "button1",
	}
	err := HandleCallback(router, query)

	assert.NoError(t, err, "Ошибка при обработке обратного вызова")
}
