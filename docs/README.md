# Library Service

## Overview
The Library Management System is a Go-based application designed to manage
authors and books. It provides functionalities to register authors,
add books, and retrieve information about authors and books.

## Run

To run, you need to set env parameters

.env:
```shell
    GRPC_PORT=9090;
    GRPC_GATEWAY_PORT=8080;
    POSTGRES_HOST=127.0.0.127;
    POSTGRES_PORT=5432;
    POSTGRES_DB=library;
    POSTGRES_USER=ed;
    POSTGRES_PASSWORD=1234567;
    POSTGRES_MAX_CONN=10;
  
    OUTBOX_ENABLED=true;
    OUTBOX_WORKERS=5;
    OUTBOX_BATCH_SIZE=100;
    OUTBOX_WAIT_TIME_MS=1000;
    OUTBOX_IN_PROGRESS_TTL_MS=1000;
    OUTBOX_AUTHOR_SEND_URL=http://book-service/send;
    OUTBOX_BOOK_SEND_URL=http://author-service/send;
```

### Docker-compose for db

```shell
    docker-compose up -d
```

```shell
    make all
    env $(cat .env | xargs) bin/library
```

## [Examples](spec/api/library/library.swagger.json)
