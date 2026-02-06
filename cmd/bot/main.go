package main

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/yandex-development-2-team/Go/internal/database"
	"github.com/yandex-development-2-team/Go/internal/logger"
	"github.com/yandex-development-2-team/Go/internal/shutdown"
	"go.uber.org/zap"
)

func main() {
	//логирование для: Запуска приложения
	env := "development"
	if env != "development" && env != "prodaction" {
		log.Fatal("Ошибка при указании метода логирования")
	}
	logger := logger.NewLogger(env)
	logger.Info("bot_started", zap.String("token_length", strconv.Itoa(len("token"))))

	//логирование для: Ошибок БД
	err := database.RunMigrations()
	if err != nil {
		logger.Error("database_error", zap.Error(err), zap.String("operation", "get_user"))
	}

	//логирование для: Входящих сообщений Telegram (только ID юзера, без content для privacy)
	logger.Info("Входящее сообщение от пользователя", zap.String("user_id", "123"))

	//логирование для: Критических ошибок
	err = fmt.Errorf("Критическая ошибка")
	logger.Fatal("Критическая ошибка", zap.Error(err))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sh := shutdown.NewShutdownHandler(logger)
	// Вызов Graceful shutdown
	go func() {
		if err := sh.WaitForShutdown(ctx, cancel,
			shutdown.ShutdownTask{
				Name: "TGUpdates",
				Fn:   updates.Stop, // Когда в будущем будет реализовано, вызов функции для остановки приема новых обновлений от TG API
			},
			shutdown.ShutdownTask{
				Name: "Database",
				Fn:   db.Close, // Когда в будущем будет реализовано, вызов функции для завершения pending queries
			},
			shutdown.ShutdownTask{
				Name: "Prometheus metrics",
				Fn:   metricsServer.Stop, // Когда в будущем будет реализовано, завершение prometheus metrics
			},
		); err != nil {
			logger.Error("Graceful shutdown completed with errors", zap.Error(err))
		}
	}()
	<-ctx.Done()
}
