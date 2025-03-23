# Определяем аргумент ENV (по умолчанию dev)
ARG ENV=dev

# Stage 1: Builder (используется только в test/prod)
FROM golang:1.24.0 AS builder
WORKDIR /app

# Загружаем зависимости (оптимизированное кеширование)
COPY src/go.mod src/go.sum ./
RUN go mod download

# Копируем код и собираем бинарник
COPY src /app
RUN go build -o services .

# Stage 2: Runtime (используется в test/prod)
FROM alpine:latest AS runtime
WORKDIR /app

# Копируем скомпилированный бинарник из builder-стадии
COPY --from=builder /app/service /app/service

# Копируем миграции и скрипт ожидания БД
COPY src/migrations /app/migrations

# Копируем общий `wait-for.sh` из `docker/`
COPY docker/wait-for.sh /app/wait-for.sh
RUN chmod +x /app/wait-for.sh

# Запускаем сервис с миграциями
CMD ["/app/wait-for.sh", "ct-postgres-${ENV}", "5432", "--", "/app/service", "migrate", "up", "--database", "DB_DSN"]

# Stage 3: Dev (используется только в dev, без сборки!)
FROM golang:1.24.0-bullseye AS dev
WORKDIR /app

# Устанавливаем необходимые системные зависимости
RUN apt update && apt install -y \
    netcat-openbsd \
    net-tools \
    && rm -rf /var/lib/apt/lists/*

# Устанавливаем пакет для миграций БД
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.2

# Пакет для .env файлов
RUN go install github.com/joho/godotenv/cmd/godotenv@v1.5.1

# Устанавливаем Delve (отладчик)
RUN go install github.com/go-delve/delve/cmd/dlv@latest

# Экспортируем PATH для доступа к go-бинарникам
ENV PATH=$PATH:/go/bin

# Копируем миграции и скрипт ожидания БД
COPY src/migrations /app/migrations

# Копируем общий `wait-for.sh` из `docker/`
COPY docker/wait-for.sh /app/wait-for.sh
RUN chmod +x /app/wait-for.sh

# Устанавливаем `air` через `go install`
RUN go install github.com/air-verse/air@v1.61.7 && \
    mv /go/bin/air /usr/local/bin/air && \
    chmod +x /usr/local/bin/air

# Копируем общий `.air.toml` из `docker/`
COPY docker/.air.toml /app/.air.toml

COPY docker/debug-wrapper.sh /app/debug-wrapper.sh
RUN chmod +x /app/debug-wrapper.sh

# Открываем порт для отладчика
EXPOSE 40000

# В dev код **не копируем** — он монтируется из локальной машины
# Используем переменную DEBUG для переключения между обычным режимом и отладкой
CMD ["/bin/sh", "-c", "/app/wait-for.sh ct-postgres-${ENV} 5432 && migrate -path /app/migrations -database \"$DB_DSN\" up && if [ \"$DEBUG\" = \"true\" ]; then /app/debug-wrapper.sh; else air -c .air.toml; fi"]

