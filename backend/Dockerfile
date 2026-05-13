# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Копируем файлы зависимостей
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
# Флаг CGO_ENABLED=0 обязателен для создания статически слинкованного бинарника для работы в alpine
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server/main.go

# Production stage
FROM alpine:latest

# Добавляем корневые сертификаты (могут понадобиться для внешних HTTP-запросов и работы с БД)
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Копируем собранный бинарник из этапа builder
COPY --from=builder /app/server .

# Открываем порт, указанный в docker-compose
EXPOSE 8080

# Запускаем приложение
CMD ["./server"]
