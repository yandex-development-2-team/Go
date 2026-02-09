run:
	go run ./cmd/bot

build:
	go build -o bin/bot ./cmd/bot

test:
	go test ./...

lint:
	golangci-lint run

migrate-create:
	goose -dir migrations create $(name) sql
