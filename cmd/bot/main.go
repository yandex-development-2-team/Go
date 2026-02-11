package main

import (
<<<<<<< HEAD
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"

	dbm "github.com/yandex-development-2-team/Go/internal/database"
)

func main() {
	log.Println("Bot starting...")

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Println("DATABASE_URL not set; skipping migrations")
		return
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	if err := dbm.RunMigrations(db); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}

	log.Println("Migrations applied successfully")
=======
	"context"
	"strings"
	"time"

	"github.com/yandex-development-2-team/Go/internal/bot"
	"github.com/yandex-development-2-team/Go/internal/config"
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
>>>>>>> 96e68a5df650fadd3caec3fbafc18e13bbc9fc93
}
