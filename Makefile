APP_NAME := existsbot
PKG := ./cmd/existsbot
OUT_DIR ?= bin

VERSION := $(or $(VERSION),dev)
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -X github.com/segfaultuwu/exists.lol/internal/version.Version=$(VERSION) \
           -X github.com/segfaultuwu/exists.lol/internal/version.Commit=$(COMMIT) \
           -X github.com/segfaultuwu/exists.lol/internal/version.BuildDate=$(BUILD_DATE)

.PHONY: all build run dev test fmt vet tidy check clean docker-build docker-up docker-down docker-logs

all: build

build:
	mkdir -p $(OUT_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/$(APP_NAME) $(PKG)

run:
	go run $(PKG)

dev:
	go run $(PKG)

test:
	go test ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

tidy:
	go mod tidy

check: fmt vet test

clean:
	rm -rf bin out

docker-build:
	docker compose build

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f existsbot
