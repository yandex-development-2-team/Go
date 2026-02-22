run:
	go run ./cmd/bot

build:
	go build -o bin/bot ./cmd/bot

test:
	go test ./...

lint:
	golangci-lint run

# Миграции
# Подгружаем переменные из .env
ifneq ($(wildcard .env),)
  include .env
  export
endif

migrate-create:
	goose -dir migrations create $(name) sql

migrate-up:
	@echo "Applying migrations..."
	goose -dir migrations postgres "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable" up

migrate-down:
	@echo "⬇Rolling back last migration..."
	goose -dir migrations postgres "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable" down

migrate-status:
	@echo "Migration status:"
	goose -dir migrations postgres "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable" status

# Docker
docker-up:
	@echo "Starting Docker containers..."
	docker-compose up -d

docker-down:
	@echo "Stopping Docker containers..."
	docker-compose down

docker-clean:
	@echo "Cleaning Docker volumes and containers..."
	docker-compose down -v

docker-logs:
	@echo "Docker logs:"
	docker-compose logs -f

# Очистка
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -rf data/  # Осторожно! Удаляет локальные данные БД
