package logger

import (
	"fmt"

	"go.uber.org/zap"
)

func NewLogger(env string) *zap.Logger {
	var (
		logger *zap.Logger
		err    error
	)

	if env == "development" {
		logger, err = zap.NewDevelopment()
		defer logger.Sync()
	} 
	if env == "production" {
		logger, err = zap.NewProduction()
		defer logger.Sync()
	} 

	if err != nil {
		panic(fmt.Sprintf("Ошибка при инициализации логгеров:%v", err))
	}

	return logger
}
