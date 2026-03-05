package handlers

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

const welcomeMessage = "👋 Добро пожаловать в Bot Яндекса!\n\nВыберите интересующую вас опцию:"

func HandleStart(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, logger *zap.Logger) error {
	if msg == nil || msg.From == nil {
		return fmt.Errorf("invalid message from user")
	}

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
