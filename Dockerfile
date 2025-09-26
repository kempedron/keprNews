# Используем Alpine версию golang для сборки бинарника
FROM golang:alpine AS builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd/myapp/main.go

# Создаем финальный образ на основе alpine
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/app .
COPY --from=builder /app/web/templates ./templates/


CMD ["./app"]