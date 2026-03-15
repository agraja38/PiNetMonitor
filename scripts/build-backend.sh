#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

echo "[PiNetMonitor] Building backend binaries"
export GOFLAGS="${GOFLAGS:-} -mod=mod"
go get github.com/mattn/go-sqlite3@v1.14.24
go build -o bin/pinetmonitord ./cmd/pinetmonitord
go build -o bin/pinetmonitor ./cmd/pinetmonitor
