FROM golang:latest

WORKDIR /application
COPY go.mod go.sum Makefile ./
RUN make bin-deps

COPY . .
RUN make .generate && (GOOS=linux GOARCH=amd64 make build)
CMD ["./bin/library"]
