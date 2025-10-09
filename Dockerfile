FROM golang:latest

WORKDIR /application
COPY . .
RUN make generate && (GOOS=linux GOARCH=amd64 make build)
CMD ["./bin/library"]