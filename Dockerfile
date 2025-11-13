FROM golang:1.23-alpine AS builder

WORKDIR /app

# 依存関係のインストール
COPY go.mod go.sum* ./
RUN go mod download

# ソースコードのコピーとビルド
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api

# 実行用の軽量イメージ
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8088

CMD ["./main"]
