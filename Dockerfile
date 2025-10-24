FROM golang:1.25.1-alpine3.22 AS builder

RUN apk update && apk add --no-cache git ca-certificates tzdata && update-ca-certificates

WORKDIR /src

# Копируем только файлы зависимостей сначала
COPY go.mod go.sum ./

# Скачиваем зависимости
RUN go mod download

# Копируем исходный код
COPY "/cmd/backend/main.go" ./
COPY src/ src/
COPY pkg/ pkg/
COPY api/ api/

# Build the binary.
RUN CGO_ENABLED=0 GOOS=linux go build -a -o ./go/bin/api ./main.go

FROM alpine:3.21

COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

WORKDIR /app

COPY --from=builder /src/go/bin/api .

CMD ["./api"]