#!/usr/bin/env bash
set -Eeuo pipefail

APP_DIR="${APP_DIR:-/app}"
LOG_DIR="${LOG_DIR:-/app/data}"
LOG_FILE="${LOG_FILE:-$LOG_DIR/self-update.log}"

MODE="${UPDATE_MODE:-docker}" # docker albo systemd
SERVICE="${SYSTEMD_SERVICE:-existsbot}"

cd "$APP_DIR"

mkdir -p "$LOG_DIR"

exec > >(tee -a "$LOG_FILE") 2>&1

echo "========================================"
echo "exists.lol self-update"
echo "time: $(date -Is)"
echo "dir: $APP_DIR"
echo "mode: $MODE"
echo "========================================"

echo "[1/7] git safe.directory"
git config --global --add safe.directory "$APP_DIR" || true

echo "[2/7] git fetch"
git fetch --all --prune

echo "[3/7] current revision"
OLD_REV="$(git rev-parse --short HEAD || echo unknown)"
echo "old rev: $OLD_REV"

echo "[4/7] pull"
git pull --ff-only

NEW_REV="$(git rev-parse --short HEAD || echo unknown)"
echo "new rev: $NEW_REV"

echo "[5/7] go mod tidy"
go mod tidy

echo "[6/7] build check"
go test ./...
go build -o bin/existsbot ./cmd/existsbot

echo "[7/7] restart"
case "$MODE" in
  docker)
    echo "restarting docker compose"
    docker compose up -d --build --force-recreate
    ;;

  systemd)
    echo "restarting systemd service: $SERVICE"
    systemctl restart "$SERVICE"
    ;;

  none)
    echo "UPDATE_MODE=none, not restarting"
    ;;

  *)
    echo "unknown UPDATE_MODE: $MODE"
    exit 1
    ;;
esac

echo "========================================"
echo "update finished"
echo "old rev: $OLD_REV"
echo "new rev: $NEW_REV"
echo "========================================"
