#!/usr/bin/env sh
set -eu

if ! command -v sqlc >/dev/null 2>&1; then
  if ! command -v go >/dev/null 2>&1; then
    echo "go is required to install sqlc" >&2
    exit 1
  fi
  go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.25.0
fi

sqlc generate
