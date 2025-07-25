# Library Service

## Overview
- [x] Реализован library-service:
  - [x] Можно добавлять книги и авторов
  - [x] Изменять книги и авторов
  - [x] Получать книги и авторов

- [x] Реализована интеграция с базой данных
  - [x] Созданы миграции с помощью goose
  
- [x] Реализован паттерн outbox

- [x] Реализован monitoring
  - [x] Для обработки логов используется loki и prometheus
  - [x] Для экспорта логов из loki используется Promtail
  - [x] Для отображения используется grafana
  - [x] Для трейсинга используется jaeger

Собираемые метрики:
### Outbox
* График количества задач
* График, показывающий скорость обработки задач
* График, показывающий rate неуспешных задач

### Library handlers
* График, показывающий количество горутин `go_goroutines`
* График, показывающий live heap `go_memstats_heap_inuse_bytes`
* График, показывающий RPS для каждого эндпоинта. 
* График, показывающий latency для каждого эндпоитна. 

### Postgres
* График, показывающий количество записей по каждой таблице в базе
* График, показывающий rate вставки записей по каждой таблице
* График, показывающий latency основных операций.

## Run

```shell
    docker-compose up -d
```

## Run

To run, you need to set env parameters

.env:
```shell
GRPC_PORT=9090
GRPC_GATEWAY_PORT=8080

POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_DB=library
POSTGRES_USER=ed
POSTGRES_PASSWORD=1234567
POSTGRES_MAX_CONN=10
METRICS_PORT=9000

JAEGER_TRACE_PORT=14268
JAEGER_WEB_PORT=16686
JAEGER_URL="http://jaeger:${JAEGER_TRACE_PORT}/api/traces"
PYROSCOPE_URL="http://pyroscope:4040"

OUTBOX_ENABLED=true
OUTBOX_WORKERS=5
OUTBOX_BATCH_SIZE=100
OUTBOX_WAIT_TIME_MS=1000
OUTBOX_IN_PROGRESS_TTL_MS=1000
OUTBOX_AUTHOR_SEND_URL="http://httpbin.org/post"
OUTBOX_BOOK_SEND_URL="http://httpbin.org/post"
```

## [Examples](spec/api/library/library.swagger.json)
