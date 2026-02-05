package config

import (
	"errors"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Telegram  TelegramConfig  `yaml:"telegram"`
	Database  DatabaseConfig  `yaml:"database"`
	Logger    LoggerConfig    `yaml:"logger"`
}

type ServerConfig struct {
	Port           int    `yaml:"port"`
	Environment    string `yaml:"environment"` // dev/prod
	PrometheusPort int    `yaml:"prometheus_port"`
}

type TelegramConfig struct {
	BotToken string `yaml:"bot_token"`
}

type DatabaseConfig struct {
	PostgresURL string `yaml:"postgres_url"`
}

type LoggerConfig struct {
	Level string `yaml:"level"`
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

	// ENV overrides

	if v := os.Getenv("BOT_TOKEN"); v != "" {
		cfg.Telegram.BotToken = v
	}

	if v := os.Getenv("POSTGRES_URL"); v != "" {
		cfg.Database.PostgresURL = v
	}

	if v := os.Getenv("PORT"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.Server.Port = i
		}
	}

	if v := os.Getenv("ENVIRONMENT"); v != "" {
		cfg.Server.Environment = v
	}

	if v := os.Getenv("PROMETHEUS_PORT"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.Server.PrometheusPort = i
		}
	}

	if v := os.Getenv("LOG_LEVEL"); v != "" {
		cfg.Logger.Level = v
	}

	// validation

	if cfg.Telegram.BotToken == "" {
		return nil, errors.New("telegram bot token is required")
	}
	if cfg.Database.PostgresURL == "" {
		return nil, errors.New("postgres url is required")
	}

	return cfg, nil
}