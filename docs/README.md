# Library Service

## Overview
The Library Management System is a Go-based application designed to manage
authors and books. It provides functionalities to register authors,
add books, and retrieve information about authors and books.

## Run

To run, need set env parameters in [.env](../.env)

### Docker-compose for db

```shell
    docker-compose up -d
```

```shell
    make all
    env $(cat .env | xargs) bin/library
```

## [Examples](spec/api/library/library.swagger.json)
