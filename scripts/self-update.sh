#!/usr/bin/env sh
set -eu

APP_DIR="/home/bots/exists.lol"
SERVICE_NAME="${SYSTEMD_SERVICE:-existsbot}"

cd "$APP_DIR"

git pull --ff-only

go mod tidy
make build

sudo /usr/bin/systemctl restart "$SERVICE_NAME"
