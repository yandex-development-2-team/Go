package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

const (
	callback           = "box_"
	callbackBackToMain = "back_to_main"
	callbackMenu       = "box_solutions"

	serviceTretyakov = 1
	servicePushkin   = 2
	serviceTheatre   = 3
	serviceTennis    = 4
	servicePadel     = 5
	serviceDigest    = 6
)

type service struct {
	id    int
	title string
}

var services = []service{
	{serviceTretyakov, "Третьяковская галерея"},
	{servicePushkin, "Пушкинский музей"},
	{serviceTheatre, "Театр на Малой Бронной"},
	{serviceTennis, "Теннис в Лужниках"},
	{servicePadel, "Падел корт в Сити"},
	{serviceDigest, "Дайджест светских событий"},
}

type BoxSolutionsHandler struct {
	bot    *tgbotapi.BotAPI
	logger *zap.Logger
}

func NewBoxSolutionsHandler(bot *tgbotapi.BotAPI, logger *zap.Logger) *BoxSolutionsHandler {
	return &BoxSolutionsHandler{
		bot:    bot,
		logger: logger,
	}
}

func (h *BoxSolutionsHandler) Handle(ctx context.Context, query *tgbotapi.CallbackQuery) error {
	data := query.Data
	userID := query.From.ID
	chatID := query.Message.Chat.ID
	messageID := query.Message.MessageID

	if data == callbackMenu {
		h.logger.Info("box_solutions_menu_opened", zap.Int64("user_id", userID), zap.Int64("chat_id", chatID))

		var rows [][]tgbotapi.InlineKeyboardButton

		for _, svc := range services {
			callbackData := fmt.Sprintf("%s%d", callback, svc.id)
			button := tgbotapi.NewInlineKeyboardButtonData(svc.title, callbackData)
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(button))
		}

		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Назад", callbackBackToMain),
		))

		keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)

		text := "📦 *Коробочные решения*\n\n" +
			"Выберите интересующее вас предложение:"

		message := tgbotapi.NewEditMessageTextAndMarkup(
			chatID,
			messageID,
			text,
			keyboard,
		)
		message.ParseMode = "Markdown"

		if _, err := h.bot.Send(message); err != nil {
			h.logger.Error("failed_to_edit_box_solutions_message", zap.Error(err), zap.Int64("user_id", userID), zap.Int("message_id", messageID))
			return err
		}
		return nil
	}

	if strings.HasPrefix(data, callback) {

		serviceIDStr := strings.TrimPrefix(data, callback)
		serviceID, err := strconv.Atoi(serviceIDStr)
		if err != nil {
			h.logger.Error("invalid_service_id", zap.String("data", data), zap.Error(err))
			return err
		}

		h.logger.Info("service_selected", zap.Int64("user_id", userID), zap.Int("service_id", serviceID))

		return HandleServiceDetail(serviceID, userID)
	}

	if data == callbackBackToMain {
		h.logger.Info("back_to_main_clicked", zap.Int64("user_id", userID))
		return nil
	}

	return nil
}
