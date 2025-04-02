# Library Service

## Overview
The Library Management System is a Go-based application designed to manage
authors and books. It provides functionalities to register authors,
add books, and retrieve information about authors and books.

## Install

```shell
    git clone git@github.com:itmo-org/ctgo-library-service-Tortik3000.git
    cd ctgo-library-service-Tortik3000
    git checkout hw
    make all
```

## Docker-compose for db

```shell
    docker-compose up -d
```

## Run

To run, you need to set env parameters in .env file
If you do not specify an environment variable, 
the default value will be selected

Default values:
```shell
    GRPC_PORT=9090
    GRPC_GATEWAY_PORT=8080
    POSTGRES_HOST=127.0.0.127
    POSTGRES_PORT=5432
    POSTGRES_DB=library
    POSTGRES_USER=ed
    POSTGRES_PASSWORD=1234567
    POSTGRES_MAX_CONN=10
```

```shell
    env $(cat .env | xargs) ../bin/library
```

## [Examples](spec/api/library/library.swagger.json)
