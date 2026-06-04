APP_NAME := existsbot
MAIN := ./cmd/existsbot
BIN_DIR := bin
BIN := $(BIN_DIR)/$(APP_NAME)

SERVICE_NAME := existsbot
SERVICE_FILE := ./services/existsbot.service
SYSTEMD_DIR := /etc/systemd/system
SYSTEM_SERVICE := $(SYSTEMD_DIR)/$(SERVICE_NAME).service

GO := go

ifneq (,$(wildcard .env))
	include .env
	export
endif

.PHONY: dev build run clean test fmt tidy install install-system uninstall restart logs status

dev:
	$(GO) run $(MAIN)

build:
	mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN) $(MAIN)

run: build
	./$(BIN)

install: install-system

install-system: build
	sudo cp $(SERVICE_FILE) $(SYSTEM_SERVICE)
	sudo systemctl daemon-reload
	sudo systemctl enable --now $(SERVICE_NAME)

uninstall:
	sudo systemctl disable --now $(SERVICE_NAME) || true
	sudo rm -f $(SYSTEM_SERVICE)
	sudo systemctl daemon-reload

restart: build
	sudo systemctl restart $(SERVICE_NAME)

logs:
	journalctl -u $(SERVICE_NAME) -f

status:
	systemctl status $(SERVICE_NAME)

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
