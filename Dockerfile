FROM golang:1.25.3-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bot ./cmd/bot/main.go


FROM alpine:latest

WORKDIR /app

# ВАЖНО для Telegram ботов:
# 1. ca-certificates — нужны для HTTPS запросов к Telegram API.
# 2. tzdata — нужна, чтобы бот понимал часовые пояса (для логов и расписаний).
RUN apk add --no-cache ca-certificates tzdata

ENV TZ=Europe/Moscow

COPY --from=builder /app/bot .

CMD ["./bot"]

