# cmd/auth-service/Dockerfile и cmd/article-service/Dockerfile
FROM golang:alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o article-service ./cmd/article-service/main.go
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/article-service .
COPY --from=builder /app/.env .
COPY --from=builder /app/web/templates ./web/templates
CMD ["./article-service"]  