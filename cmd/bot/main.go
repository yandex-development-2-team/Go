package main

import (
	"context"
	"database/sql"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/yandex-development-2-team/Go/internal/bot"
	"github.com/yandex-development-2-team/Go/internal/config"
	"github.com/yandex-development-2-team/Go/internal/database"
	"github.com/yandex-development-2-team/Go/internal/logger"
	"github.com/yandex-development-2-team/Go/internal/shutdown"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	// internal/logger.NewLogger ожидает "development" / "production"
	env := "development"
	if strings.ToLower(cfg.Server.Environment) == "prod" {
		env = "production"
	}

	log := logger.NewLogger(env)
	defer func() { _ = log.Sync() }()

	db, err := sql.Open("postgres", cfg.Database.PostgresURL)
	if err != nil {
		log.Fatal("failed_to_open_db", zap.Error(err))
	}
	defer func() { _ = db.Close() }()

	if err := db.Ping(); err != nil {
		log.Fatal("failed_to_ping_db", zap.Error(err))
	}

	if err := database.RunMigrations(db); err != nil {
		log.Fatal("failed_to_run_migrations", zap.Error(err))
	}

	tg, err := bot.NewTelegramBot(cfg.Telegram.BotToken, log)
	if err != nil {
		log.Fatal("failed_to_init_bot", zap.Error(err))
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	updates, err := tg.GetUpdates(ctx, 30*time.Second)
	if err != nil {
		log.Fatal("failed_to_get_updates", zap.Error(err))
	}

	// graceful shutdown: по SIGINT/SIGTERM отменяем контекст
	sh := shutdown.NewShutdownHandler(log)
	go func() {
		if err := sh.WaitForShutdown(ctx, cancel); err != nil {
			log.Error("Graceful shutdown completed with errors", zap.Error(err))
		}
	}()

	for range updates {
		// обработчики добавятся позже; важно лишь, что polling работает и не падает
	}
}
