# Установим базовый образ
FROM golang:1.24.2-alpine AS builder

# Установим рабочую директорию
WORKDIR /app

# Указываем путь к файлам go.mod и go.sum через build-args
ARG GO_MOD_PATH=./go.mod
ARG GO_SUM_PATH=./go.sum

# Копируем go.mod и go.sum с указанным путем
COPY ${GO_MOD_PATH} go.mod
COPY ${GO_SUM_PATH} go.sum

# Устанавливаем зависимости
RUN go mod download

# Копируем весь проект в контейнер
COPY . .

COPY .env .env

# Строим приложение
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/marketflow ./cmd/marketflow

# Создаем новый минималистичный контейнер для запуска приложения
FROM alpine:latest

# Устанавливаем зависимости, если они нужны (например, для работы с SSL)
RUN apk --no-cache add ca-certificates

# Копируем собранное приложение из первого этапа
COPY --from=builder /bin/marketflow /bin/marketflow

# Экспонируем порт, который использует ваше приложение
EXPOSE 8080

# Запускаем приложение
CMD ["/bin/marketflow"]