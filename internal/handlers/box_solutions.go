package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yandex-development-2-team/Go/internal/database/repository"
	"go.uber.org/zap"
)

const (
	callback     = "box_"
	callbackMenu = "box_solutions"
)

type BoxSolutionsHandler struct {
	bot      *tgbotapi.BotAPI
	logger   *zap.Logger
	services *repository.ServiceRepository
}

func NewBoxSolutionsHandler(bot *tgbotapi.BotAPI, logger *zap.Logger, services *repository.ServiceRepository) *BoxSolutionsHandler {
	return &BoxSolutionsHandler{
		bot:      bot,
		logger:   logger,
		services: services,
	}
}

func (h *BoxSolutionsHandler) Handle(ctx context.Context, query *tgbotapi.CallbackQuery) error {
	data := query.Data
	userID := query.From.ID
	chatID := query.Message.Chat.ID
	messageID := query.Message.MessageID

	if data == callbackMenu {
		h.logger.Info("box_solutions_menu_opened", zap.Int64("user_id", userID), zap.Int64("chat_id", chatID))

		services, err := h.services.GetServicesOfBoxSolutions(ctx)
		if err != nil {
			h.logger.Error("failed_to_get_services", zap.Error(err))
			return err
		}

		var rows [][]tgbotapi.InlineKeyboardButton

		for _, svc := range services {
			callbackData := fmt.Sprintf("%s%d", callback, svc.ID)
			button := tgbotapi.NewInlineKeyboardButtonData(svc.Title, callbackData)
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(button))
		}

		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Назад", CallbackBackToMain),
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

	if data == CallbackBackToMain {
		return NewMainMenuHandler(h.bot, h.logger).Handle(ctx, query)
	}

	return nil
}
