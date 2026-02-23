FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata build-base

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bot ./cmd/bot

FROM alpine:3.20

# ВАЖНО: установить CA-сертификаты!
RUN apk --no-cache add ca-certificates tzdata && \
    update-ca-certificates 2>/dev/null || true

COPY --from=builder /bot /bot

WORKDIR /app

# Запуск
CMD ["/bot"]