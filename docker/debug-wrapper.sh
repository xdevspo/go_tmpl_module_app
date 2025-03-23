#!/bin/sh
# Компилируем с флагами отладки
go build -gcflags="all=-N -l" -o ./tmp/main ./cmd/main.go

# Запускаем отладчик
exec dlv exec --headless --listen=0.0.0.0:40000 --api-version=2 --accept-multiclient --continue ./tmp/main