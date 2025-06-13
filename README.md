# Сервис статистики баннеров

Сервис на Go для управления и отслеживания статистики баннеров. Предоставляет API для работы с баннерами и их статистикой.

## Возможности

- REST API для управления баннерами
- Интеграция с PostgreSQL
- Поддержка Docker для простого развертывания
- Структурированное логирование с использованием zerolog
- Корректное завершение работы
- Утилита нагрузочного тестирования

## Требования

- Go 1.24
- Docker и Docker Compose
- PostgreSQL (при локальном запуске)

## Конфигурация

Сервис можно настроить с помощью переменных окружения:

- `PORT` - Порт сервера (по умолчанию: 3000)
- `DATABASE_URL` - Строка подключения к PostgreSQL (по умолчанию: postgres://postgres:postgres@localhost:5432/banners?sslmode=disable)

## Запуск сервиса

### Использование Docker

1. Сборка и запуск с помощью Docker Compose:
```bash
docker-compose up --build
```

### Локальный запуск

1. Установка зависимостей:
```bash
go mod download
```

2. Запуск сервиса:
```bash
go run cmd/main.go
```

## Нагрузочное тестирование

Для проведения нагрузочного тестирования сервиса используйте утилиту в директории `cmd/loadtest`:
```bash
go run cmd/loadtest/main.go [flags]
```

### Флаги нагрузочного тестирования

- `-url` - базовый URL тестируемого сервиса (по умолчанию "http://localhost:3000")
- `-requests` - общее количество запросов для каждого эндпоинта (по умолчанию 1000)
- `-concurrency` - количество одновременных запросов (по умолчанию 10)
- `-banner` - ID баннера для тестирования (по умолчанию 5)

### Пример запуска теста
```bash
go run cmd/loadtest/main.go -url="http://localhost:3000" -requests=1000 -concurrency=10 -banner=5
```

### Вывод результатов тестирования

Утилита выводит статистику отдельно для каждого эндпоинта:
```aiignore
Click Endpoint Results:
Total Requests: 1000
Successful Requests: 1000
Failed Requests: 0
RPS: 814.28
Min Latency: 1.597375ms
Max Latency: 46.211375ms
Average Latency: 8.117849ms

Stats Endpoint Results:
Total Requests: 1000
Successful Requests: 1000
Failed Requests: 0
RPS: 814.25
Min Latency: 1.502792ms
Max Latency: 21.228875ms
Average Latency: 4.144466ms

Total Test Duration: 1.228238584s
```