#!/usr/bin/env sh
set -eu

APP_DIR="${APP_DIR:-/app}"
MODE="${UPDATE_MODE:-docker}"
SERVICE="${SYSTEMD_SERVICE:-existsbot}"

echo "========================================"
echo "exists.lol self-update"
echo "time: $(date -Is)"
echo "dir: $APP_DIR"
echo "mode: $MODE"
echo "========================================"

if [ ! -d "$APP_DIR" ]; then
  echo "APP_DIR does not exist: $APP_DIR"
  exit 1
fi

cd "$APP_DIR"

echo "[1/6] git safe.directory"
git config --global --add safe.directory "$APP_DIR" || true

echo "[2/6] current revision"
OLD_REV="$(git rev-parse --short HEAD 2>/dev/null || echo unknown)"
echo "old rev: $OLD_REV"

echo "[3/6] git fetch/pull"
git fetch --all --prune
git pull --ff-only

NEW_REV="$(git rev-parse --short HEAD 2>/dev/null || echo unknown)"
echo "new rev: $NEW_REV"

echo "[4/6] mode-specific update"

case "$MODE" in
  docker)
    echo "docker mode: rebuilding via docker compose"

    if ! command -v docker >/dev/null 2>&1; then
      echo "docker command not found inside container"
      echo "Install docker-cli in Dockerfile and mount /var/run/docker.sock"
      exit 1
    fi

    docker compose up -d --build --force-recreate
    ;;

  systemd)
    echo "systemd mode: building locally"

    if ! command -v go >/dev/null 2>&1; then
      echo "go command not found"
      exit 1
    fi

    go mod tidy
    go test ./...
    mkdir -p bin
    go build -o bin/existsbot ./cmd/existsbot

    systemctl restart "$SERVICE"
    ;;

  none)
    echo "UPDATE_MODE=none, skipping restart/build"
    ;;

  *)
    echo "unknown UPDATE_MODE: $MODE"
    exit 1
    ;;
esac

echo "[5/6] final revision"
echo "old rev: $OLD_REV"
echo "new rev: $NEW_REV"

echo "[6/6] done"
echo "self-update finished"
