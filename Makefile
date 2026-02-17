run:
	go run ./cmd/bot

build:
	go build -o bin/bot ./cmd/bot

test:
	go test ./...

lint:
	golangci-lint run

MIGRATIONS_DIR ?= migrations
GOOSE := go run github.com/pressly/goose/v3/cmd/goose@v3.26.0

.PHONY: migrate-status migrate-up migrate-down migrate-redo migrate-reset migrate-create

migrate-status:
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(POSTGRES_URL)" status

migrate-up:
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(POSTGRES_URL)" up

migrate-down:
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(POSTGRES_URL)" down

migrate-redo:
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(POSTGRES_URL)" redo

migrate-reset:
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(POSTGRES_URL)" reset

# Создание новой SQL миграции с timestamp-префиксом
# пример: make migrate-create name=create_users_table
migrate-create:
	$(GOOSE) -dir $(MIGRATIONS_DIR) create $(name) sql
