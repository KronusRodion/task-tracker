# Стадия сборки
FROM golang:1.26.4-alpine3.24 AS builder

# Установка рабочей директории
WORKDIR /app

# Копирование зависимостей
COPY go.mod go.sum ./

# Загрузка зависимостей
RUN go mod download

# Копирование исходного кода
COPY . .

# Сборка приложения
RUN go build -o /app/service ./cmd/task-tracker/main.go

# Стадия финального образа
FROM alpine:3.23

ENV TZ=UTC \
    APP_USER=appuser \
    APP_GROUP=appgroup \
    APP_HOME=/app

# Создание непривилегированного пользователя
RUN addgroup -g 1000 -S ${APP_GROUP} \
    && adduser -u 1000 -S ${APP_USER} -G ${APP_GROUP} -h ${APP_HOME}

USER ${APP_USER}

# Создание рабочей директории
WORKDIR /app

# Копирование скомпилированного бинарного файла
COPY --from=builder /app/service /app/service
EXPOSE 8080
# Команда для запуска приложения
CMD ["/app/service"]