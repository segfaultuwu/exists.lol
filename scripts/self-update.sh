#!/usr/bin/env sh
set -eu

APP_DIR="/home/segfault/exists.lol"
SERVICE_NAME="${SYSTEMD_SERVICE:-existsbot}"

cd "$APP_DIR"

git pull --ff-only

go mod tidy
make build

systemctl --user restart "$SERVICE_NAME"
