package handlers

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

const (
	CallbackBackToMain   = "back_to_main"
	CallbackBoxSolutions = "box_solutions"
)

type MainMenuHandler struct {
	bot    *tgbotapi.BotAPI
	logger *zap.Logger
}

func NewMainMenuHandler(bot *tgbotapi.BotAPI, logger *zap.Logger) *MainMenuHandler {
	return &MainMenuHandler{
		bot:    bot,
		logger: logger,
	}
}

func (h *MainMenuHandler) Handle(ctx context.Context, query *tgbotapi.CallbackQuery) error {
	userID := query.From.ID
	chatID := query.Message.Chat.ID
	messageID := query.Message.MessageID

	text := "🏠 *Главное меню*\n\n" +
		"Выберите интересующий вас раздел:"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📦 Коробочные решения", CallbackBoxSolutions),
		),
	)

	message := tgbotapi.NewEditMessageTextAndMarkup(
		chatID,
		messageID,
		text,
		keyboard,
	)
	message.ParseMode = "Markdown"

	if _, err := h.bot.Send(message); err != nil {
		h.logger.Error("failed_to_open_main_menu", zap.Error(err), zap.Int64("user_id", userID), zap.Int("message_id", messageID))
		return err
	}

	h.logger.Info("main_menu_opened", zap.Int64("user_id", userID), zap.Int64("chat_id", chatID))
	return nil
}
