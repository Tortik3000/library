# Документация сервиса Библиотека

## Обзор проекта

**Библиотека** — это микросервис на языке Go для управления книгами и авторами, реализующий современную архитектуру с поддержкой REST API, gRPC, мониторинга, трейсинга и профилирования.

---

## Основные возможности

### Управление книгами
- Добавление новой книги (`POST /v1/library/book`)
- Обновление существующей книги (`PUT /v1/library/book`)
- Получение информации о книге по ID (`GET /v1/library/book/{id}`)

### Управление авторами
- Регистрация нового автора (`POST /v1/library/author`)
- Обновление информации об авторе (`PUT /v1/library/author`)
- Получение информации об авторе по ID (`GET /v1/library/author/{id}`)
- Получение всех книг конкретного автора (`GET /v1/library/author_books/{author_id}`)

### Валидация
- Идентификаторы — UUID, генерируемые автоматически БД
- Имена авторов: регулярное выражение `^[A-Za-z0-9]+( [A-Za-z0-9]+)*$`, длина 1-512 символов
- Валидация на уровне protobuf через `protoc-gen-validate`

---

## Технологический стек

### Основные технологии
- **Язык**: Go 1.25
- **API**: gRPC с REST Gateway
- **База данных**: PostgreSQL 17 с драйвером [pgx](https://github.com/jackc/pgx)
- **Миграции**: [goose](https://github.com/pressly/goose)

### Observability Stack
- **Метрики**: [Prometheus](https://prometheus.io/) + [Grafana](https://grafana.com/)
- **Логирование**: [Loki](https://grafana.com/oss/loki/) + [Promtail](https://grafana.com/docs/loki/latest/clients/promtail/)
- **Трейсинг**: [Jaeger](https://www.jaegertracing.io/)
- **Профилирование**: [Pyroscope](https://grafana.com/oss/pyroscope/)

### Дополнительные инструменты
- **Валидация**: [protoc-gen-validate](https://github.com/bufbuild/protoc-gen-validate)
- **Логирование (код)**: [logrus](https://github.com/sirupsen/logrus)
- **Контейнеризация**: Docker, Docker Compose

---

## API Documentation

### Swagger Specification

Полная спецификация API доступна в формате Swagger:

**Локальный файл**: [`docs/spec/api/library/library.swagger.json`](spec/api/library/library.swagger.json)

---

## Мониторинг и метрики

Сервис собирает и экспортирует метрики в Prometheus, которые визуализируются в Grafana через предустановленные дашборды.

### Категории метрик

#### 1. **Outbox Worker Metrics**
Метрики для мониторинга асинхронной обработки событий:

| Метрика | Тип | Описание |
|---------|-----|----------|
| `library_service_outbox_tasks_created_total` | Counter | Общее число созданных задач по типу (kind) |
| `library_service_outbox_tasks_processed_total` | Counter | Успешно обработанные задачи |
| `library_service_outbox_tasks_failed_total` | Counter | Задачи, завершившиеся ошибкой |
| `library_service_outbox_task_processing_duration_seconds` | Histogram | Время обработки задачи |

**Дашборды**:
- График количества задач в очереди
- Скорость обработки задач (tasks/sec)
- Rate неуспешных задач

#### 2. **gRPC Handler Metrics**
Метрики HTTP/gRPC эндпоинтов:

| Метрика | Тип | Описание |
|---------|-----|----------|
| `library_service_grpc_requests_total` | Counter | Общее число запросов по service/method/code |
| `library_service_grpc_request_duration_seconds` | Histogram | Время обработки запроса |

**Дашборды**:
- RPS (Requests Per Second) для каждого эндпоинта
- Latency (P50, P95, P99) для каждого эндпоинта
- Коды ответов (распределение 2xx, 4xx, 5xx)

#### 3. **Go Runtime Metrics**
Стандартные метрики Go приложения:

| Метрика | Описание |
|---------|----------|
| `go_goroutines` | Количество активных горутин |
| `go_memstats_heap_inuse_bytes` | Используемая память heap |
| `go_memstats_alloc_bytes` | Выделенная память |
| `go_gc_duration_seconds` | Длительность сборки мусора |

**Дашборды**:
- График количества горутин
- Live heap memory usage
- GC паузы и частота

#### 4. **PostgreSQL Metrics**
Метрики операций с базой данных:

| Метрика | Тип | Описание |
|---------|-----|----------|
| `library_service_db_table_num_rows` | Gauge | Количество записей по таблицам |
| `library_service_db_table_insert_total` | Counter | Общее число вставок по таблицам |
| `library_service_db_query_latency_seconds` | Histogram | Latency операций БД |

**Дашборды**:
- Количество записей по таблицам (author, book, author_book, outbox)
- Rate вставки записей
- Latency операций (SELECT, INSERT, UPDATE, DELETE)

### Доступ к Grafana

После запуска через Docker Compose:
- **URL**: http://localhost:3000
- **Логин**: `admin`
- **Пароль**: `admin`

Дашборды автоматически загружаются из директории `infra/grafana/dashboards/`.

---

## Запуск проекта

### Требования

Для работы необходимо установить:
- [Docker](https://www.docker.com/) версии 20.10+
- [Docker Compose](https://docs.docker.com/compose/) версии 1.29+
- [Make](https://www.gnu.org/software/make/) (опционально, для разработки)

### Настройка переменных окружения

[.env](../.env) файл:

```
# Порты сервиса
GRPC_PORT=9070
GRPC_GATEWAY_PORT=8080
METRICS_PORT=9000

# PostgreSQL
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_DB=library
POSTGRES_USER=ed
POSTGRES_PASSWORD=1234567
POSTGRES_MAX_CONN=10

# Jaeger (трейсинг)
JAEGER_TRACE_PORT=14268
JAEGER_WEB_PORT=16686
JAEGER_URL="http://jaeger:${JAEGER_TRACE_PORT}/api/traces"

# Pyroscope (профилирование)
PYROSCOPE_PORT=4040
PYROSCOPE_URL="http://pyroscope:${PYROSCOPE_PORT}"

# Grafana & Prometheus
GRAFANA_PORT=3000
PROMETHEUS_PORT=9090
LOKI_PORT=3100
DS_PROMETHEUS=P7847DFF4E1A49A3A

# Outbox Worker
OUTBOX_ENABLED=true
OUTBOX_WORKERS=5
OUTBOX_BATCH_SIZE=100
OUTBOX_WAIT_TIME_MS=1000
OUTBOX_IN_PROGRESS_TTL_MS=1000
OUTBOX_AUTHOR_SEND_URL="http://dummy-author:8081"
OUTBOX_BOOK_SEND_URL="http://dummy-book:8082"
```


### Запуск через Docker Compose

1. **Запустите все сервисы**:
```shell script
  docker-compose up -d
```

### Доступные сервисы

После успешного запуска доступны следующие интерфейсы:

| Сервис | URL | Описание |
|--------|-----|----------|
| REST API | http://localhost:8080 | REST эндпоинты библиотеки |
| gRPC | localhost:9070 | gRPC сервер |
| Prometheus | http://localhost:9090 | Метрики |
| Grafana | http://localhost:3000 | Дашборды (admin/admin) |
| Jaeger | http://localhost:16686 | Трейсы запросов |
| Pyroscope | http://localhost:4040 | Профилирование |

### Остановка и очистка

```shell script
# Остановка сервисов
docker-compose down

# Остановка с удалением volumes (данные БД)
docker-compose down -v

# Пересборка образов
docker-compose up -d --build
```


---


### Тестирование

```shell script
# Все тесты
make test

# Линтер
make lint

# Полный цикл (lint + test)
make all
```


---

## Структура проекта

```
library/
├── api/                    # Protobuf определения
├── cmd/                    # Точки входа приложения
├── config/                 # Конфигурация
├── db/                     # Миграции и скрипты БД
├── docs/                   # Документация
│   └── spec/              # Swagger спецификации
├── generated/             # Сгенерированный код (proto, mocks)
├── infra/                 # Конфигурация инфраструктуры
│   ├── grafana/          # Дашборды и datasources
│   ├── prometheus.yml    # Конфигурация Prometheus
│   └── promtail-config.yaml
├── integration-test/      # Интеграционные тесты
├── internal/              # Внутренний код приложения
│   ├── controller/       # HTTP/gRPC хендлеры
│   ├── usecase/          # Бизнес-логика
│   ├── repository/       # Работа с БД
│   ├── metrics/          # Prometheus метрики
│   └── ...
├── k6/                    # Нагрузочные тесты
├── .env                   # Переменные окружения
├── docker-compose.yml     # Оркестрация сервисов
├── Dockerfile             # Образ приложения
├── Makefile               # Задачи сборки
└── go.mod                 # Go зависимости
```

---

## Нагрузочное тестирование

### k6

Проект включает набор нагрузочных тестов, написанных с использованием [k6](https://k6.io/) — современного инструмента для тестирования производительности и нагрузки.

### Установка k6

Перед запуском тестов установите k6:


```shell
    sudo apt-get update
    sudo apt-get install k6
```

Для запуска тестов:
```shell
    k6 run k6/parallel_endpoint_load.js
    k6 run k6/load_test.js
```
---

## Ссылки

- **Swagger API**: [library.swagger.json](spec/api/library/library.swagger.json)
- **Protobuf**: [library.proto](../api/library/library.proto)
- **go-clean-template**: https://github.com/evrone/go-clean-template
