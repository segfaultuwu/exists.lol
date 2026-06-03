APP_NAME := existsbot
MAIN := ./cmd/existsbot
BIN_DIR := bin
BIN := $(BIN_DIR)/$(APP_NAME)

GO := go

ifneq (,$(wildcard .env))
	include .env
	export
endif

.PHONY: dev build run clean test fmt tidy

dev:
	$(GO) run $(MAIN)

build:
	mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN) $(MAIN)

run: build
	./$(BIN)

clean:
	rm -rf $(BIN_DIR)
	rm -f coverage.out
	$(GO) clean

test:
	$(GO) test ./...

fmt:
	$(GO) fmt ./...

tidy:
	$(GO) mod tidy
