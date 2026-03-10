package handlers

import (
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yandex-development-2-team/Go/internal/metrics"
	"go.uber.org/zap"
)

const welcomeMessage = "👋 Добро пожаловать в Bot Яндекса!\n\nВыберите интересующую вас опцию:"

func HandleStart(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, logger *zap.Logger) error {
	if msg == nil || msg.From == nil {
		return fmt.Errorf("invalid message from user")
	}

	start := time.Now()
	// ActiveUsers: считаем “активных” как количество одновременно обрабатываемых запросов
	metrics.Default.ActiveUsers.Inc()
	defer metrics.Default.ActiveUsers.Dec()

	metrics.Default.MessagesReceived.Inc()

	var handlerErr error
	defer func() {
		dur := time.Since(start).Seconds()
		metrics.Default.MessageProcessingDuration.Observe(dur)

		logger.Info("handle_start_metrics",
			zap.Int64("user_id", msg.From.ID),
			zap.Float64("duration_seconds", dur),
			zap.Bool("success", handlerErr == nil),
		)
	}()

	user := msg.From

	logger.Info("received start command",
		zap.Int64("user_id", user.ID),
		zap.String("username", user.UserName),
		zap.Int64("chat_id", msg.Chat.ID),
	)

	//формируем кнопки
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Коробочные решения", "box_solutions"),
			tgbotapi.NewInlineKeyboardButtonData("Гайд по посещению", "visit_guide"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Запрос спецпроекта", "special_project"),
			tgbotapi.NewInlineKeyboardButtonData("Примеры спецпроектов", "project_examples"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("О нас", "about_us"),
			tgbotapi.NewInlineKeyboardButtonData("Связь с поддержкой", "support"),
		),
	)

	//отправляем сообщение
	message := tgbotapi.NewMessage(msg.Chat.ID, welcomeMessage)
	message.ParseMode = "HTML"
	message.ReplyMarkup = inlineKeyboard

	//обрабатываем ошибку отправки
	if _, err := bot.Send(message); err != nil {
		handlerErr = err
		metrics.Default.MessagesErrorsTotal.Inc()

		logger.Error("failed to send message",
			zap.Int64("user_id", user.ID),
			zap.Int64("chat_id", msg.Chat.ID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to send start message: %w", err)
	}

	logger.Info("start message sent",
		zap.Int64("user_id", user.ID),
		zap.Int64("chat_id", msg.Chat.ID),
	)

	return nil
}
