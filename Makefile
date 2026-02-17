run:
	go run ./cmd/bot

build:
	go build -o bin/bot ./cmd/bot

test:
	go test ./...

lint:
	golangci-lint run

# Подгружаем переменные из .env (если файл есть)
ifneq ($(wildcard .env),)
include .env
export
endif

MIGRATIONS_DIR ?= migrations
GOOSE := go run github.com/pressly/goose/v3/cmd/goose@v3.26.0

.PHONY: migrate-status migrate-up migrate-down migrate-redo migrate-reset migrate-create \
	docker-up docker-down docker-clean docker-logs clean

migrate-status:
	@test -n "$(POSTGRES_URL)" || (echo "POSTGRES_URL is required"; exit 1)
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(POSTGRES_URL)" status

migrate-up:
	@test -n "$(POSTGRES_URL)" || (echo "POSTGRES_URL is required"; exit 1)
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(POSTGRES_URL)" up

migrate-down:
	@test -n "$(POSTGRES_URL)" || (echo "POSTGRES_URL is required"; exit 1)
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(POSTGRES_URL)" down

migrate-redo:
	@test -n "$(POSTGRES_URL)" || (echo "POSTGRES_URL is required"; exit 1)
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(POSTGRES_URL)" redo

migrate-reset:
	@test -n "$(POSTGRES_URL)" || (echo "POSTGRES_URL is required"; exit 1)
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(POSTGRES_URL)" reset

# Создание новой SQL миграции с timestamp-префиксом
# пример: make migrate-create name=create_users_table
migrate-create:
	$(GOOSE) -dir $(MIGRATIONS_DIR) create $(name) sql

# Docker
docker-up:
	@echo "Starting Docker containers..."
	docker compose up -d

docker-down:
	@echo "Stopping Docker containers..."
	docker compose down

docker-clean:
	@echo "Cleaning Docker volumes and containers..."
	docker compose down -v

docker-logs:
	@echo "Docker logs:"
	docker compose logs -f

# Очистка
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -rf data/  # Осторожно! Удаляет локальные данные БД