# Этап 1: Сборка
FROM golang:1.23 AS builder

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем файлы go.mod и go.sum (если есть)
COPY go.mod go.sum ./

# Скачиваем зависимости
RUN go mod download

# Копируем остальные файлы проекта
COPY . .

# Собираем статически связанный бинарник
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/bookcrossing

# Этап 2: Запуск
FROM alpine:3.19

WORKDIR /app

# Устанавливаем сертификаты для HTTPS запросов
RUN apk --no-cache add ca-certificates

# Копируем бинарник из этапа сборки
COPY --from=builder /app/main .

# Указываем порт, который будет использовать приложение
EXPOSE 8080

# Команда для запуска приложения
CMD ["./main"]