APP_NAME := existsbot
MAIN := ./cmd/existsbot
BIN_DIR := bin
BIN := $(BIN_DIR)/$(APP_NAME)

SERVICE_NAME := existsbot
SERVICE_FILE := ./services/existsbot.service
USER_SYSTEMD_DIR := $(HOME)/.config/systemd/user
INSTALLED_SERVICE := $(USER_SYSTEMD_DIR)/$(SERVICE_NAME).service

GO := go

ifneq (,$(wildcard .env))
	include .env
	export
endif

.PHONY: dev build run clean test fmt tidy install uninstall restart logs status

dev:
	$(GO) run $(MAIN)

build:
	mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN) $(MAIN)

run: build
	./$(BIN)

install: build
	mkdir -p $(USER_SYSTEMD_DIR)
	cp $(SERVICE_FILE) $(INSTALLED_SERVICE)
	systemctl --user daemon-reload
	systemctl --user enable --now $(SERVICE_NAME)

uninstall:
	systemctl --user disable --now $(SERVICE_NAME) || true
	rm -f $(INSTALLED_SERVICE)
	systemctl --user daemon-reload

restart: build
	systemctl --user restart $(SERVICE_NAME)

logs:
	journalctl --user -u $(SERVICE_NAME) -f

status:
	systemctl --user status $(SERVICE_NAME)

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
