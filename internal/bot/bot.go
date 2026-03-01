package bot

import (
	"context"
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

type TelegramBot struct {
	Api    *tgbotapi.BotAPI
	logger *zap.Logger
}

func NewTelegramBot(token string, logger *zap.Logger) (*TelegramBot, error) {
	if logger == nil {
		logger = zap.NewNop()
	}
	if token == "" {
		return nil, fmt.Errorf("telegram token is empty")
	}

	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram bot api: %w", err)
	}

	me, err := api.GetMe()
	if err != nil {
		return nil, fmt.Errorf("failed to validate token via getMe: %w", err)
	}

	logger.Info("telegram_bot_started",
		zap.Int64("bot_id", me.ID),
		zap.String("bot_username", me.UserName),
		zap.String("bot_first_name", me.FirstName),
	)

	return &TelegramBot{
		Api:    api,
		logger: logger,
	}, nil
}

func (bot *TelegramBot) GetUpdates(ctx context.Context, timeout time.Duration) (tgbotapi.UpdatesChannel, error) {
	if bot == nil || bot.Api == nil {
		return nil, fmt.Errorf("telegram bot api is nil")
	}
	if ctx == nil {
		return nil, fmt.Errorf("context is nil")
	}
	if timeout <= 0 {
		timeout = time.Second * 30
	}

	cfg := tgbotapi.NewUpdate(0)
	cfg.Timeout = int(timeout.Seconds())
	cfg.AllowedUpdates = []string{
		"message",
		"callback_query",
		"my_chat_member",
	}

	updates := bot.Api.GetUpdatesChan(cfg)

	go func() {
		<-ctx.Done()
		bot.Api.StopReceivingUpdates()
	}()

	return updates, nil
}

func (bot *TelegramBot) Ping(ctx context.Context) error {
	if bot == nil || bot.api == nil {
		return fmt.Errorf("telegram bot api is nil")
	}

	errCh := make(chan error, 1)
	go func() {
		_, err := bot.api.GetMe()
		errCh <- err
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}
