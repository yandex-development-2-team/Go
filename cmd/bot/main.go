package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/yandex-development-2-team/Go/internal/database"
	"github.com/yandex-development-2-team/Go/internal/logger"
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
}
