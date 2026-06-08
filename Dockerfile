FROM golang:1.22-alpine AS builder
RUN apk add --no-cache ca-certificates tzdata && \
    adduser -D -u 1001 appuser
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /bin/bot ./cmd/bot

FROM alpine:3.19
LABEL org.opencontainers.image.source="https://github.com/sunba91-su/Roket.Chat-GeekBot"
LABEL org.opencontainers.image.description="Rocket.Chat daily standup bot"
LABEL org.opencontainers.image.licenses="MIT"

RUN apk add --no-cache ca-certificates tzdata && \
    adduser -D -u 1001 appuser

COPY --from=builder /bin/bot /bot

USER appuser
WORKDIR /home/appuser
VOLUME /data
ENV STANDUP_DB_PATH=/data/standup-bot.db

CMD ["/bot"]
