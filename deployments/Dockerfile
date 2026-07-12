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
    app_USER=appuser \
    app_GROUP=appgroup \
    app_HOME=/app

# Создание непривилегированного пользователя
RUN addgroup -g 1000 -S ${app_GROUP} \
    && adduser -u 1000 -S ${app_USER} -G ${app_GROUP} -h ${app_HOME}



# Создание рабочей директории
WORKDIR /app

# Копирование скомпилированного бинарного файла и конфига
COPY --from=builder /app/service /app/service
COPY --from=builder /app/configs/config.yaml /app/config.yaml

RUN chown -R ${app_USER}:${app_GROUP} /app
USER ${app_USER}

EXPOSE 8080
# Команда для запуска приложения
CMD ["/app/service"]