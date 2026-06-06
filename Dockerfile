FROM golang:1.26-alpine AS builder

WORKDIR /src

RUN apk add --no-cache git make

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN make build OUT_DIR=/bin

FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates git make wget docker-cli docker-cli-compose && \
    git config --global --add safe.directory /app

COPY --from=builder /bin/existsbot /app/existsbot

RUN mkdir -p /app/domains /app/data /app/scripts

EXPOSE 8080

CMD ["/app/existsbot"]
