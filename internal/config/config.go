package config

import (
	"errors"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	TelegramBotToken string `yaml:"telegram_bot_token"`
	PostgresURL      string `yaml:"postgres_url"`
	Port             int    `yaml:"port"`
	Environment      string `yaml:"environment"` // dev/prod
	PrometheusPort   int    `yaml:"prometheus_port"`
	LogLevel         string `yaml:"log_level"`
}

func LoadConfig() (*Config, error) {
	cfg := &Config{}

	if _, err := os.Stat("config/config.yaml"); err == nil {
		data, err := os.ReadFile("config/config.yaml")
		if err != nil {
			return nil, err
		}

		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, err
		}
	}

	if v := os.Getenv("BOT_TOKEN"); v != "" {
		cfg.TelegramBotToken = v
	}

	if v := os.Getenv("POSTGRES_URL"); v != "" {
		cfg.PostgresURL = v
	}

	if v := os.Getenv("PORT"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.Port = i
		}
	}

	if v := os.Getenv("ENVIRONMENT"); v != "" {
		cfg.Environment = v
	}

	if v := os.Getenv("PROMETHEUS_PORT"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.PrometheusPort = i
		}
	}

	if v := os.Getenv("LOG_LEVEL"); v != "" {
		cfg.LogLevel = v
	}

	if cfg.TelegramBotToken == "" {
		return nil, errors.New("telegram bot token is required")
	}
	if cfg.PostgresURL == "" {
		return nil, errors.New("postgres url is required")
	}

	return cfg, nil
}
