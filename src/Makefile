# Переменная версии (по умолчанию `v1`)
VERSION ?= v1

# Переменная окружения (по умолчанию `dev`)
ENV ?= dev
ENV_FILE = .env.$(ENV)
MIGRATION_NAME ?= $(shell date +%Y%m%d%H%M%S)
MIGRATION_PATH ?= 00000

# Создание миграции
migrate-create:
	set -a && export $$(grep -v '^#' $(ENV_FILE) | xargs) && set +a && \
	migrate create -ext sql -dir ./migrations -seq $(MIGRATION_NAME)

# Миграции UP
migrate-up:
	set -a && export $$(grep -v '^#' $(ENV_FILE) | xargs) && set +a && \
    migrate -path ./migrations -database "postgres://$$POSTGRES_USER:$$POSTGRES_PASSWORD@$$POSTGRES_HOST:$$POSTGRES_PORT/$$POSTGRES_DB?sslmode=disable" up

# Миграции DOWN
migrate-down:
	set -a && export $$(grep -v '^#' $(ENV_FILE) | xargs) && set +a && \
    migrate -path ./migrations -database "postgres://$$POSTGRES_USER:$$POSTGRES_PASSWORD@$$POSTGRES_HOST:$$POSTGRES_PORT/$$POSTGRES_DB?sslmode=disable" down