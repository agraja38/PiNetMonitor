#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

echo "[PiNetMonitor] Building backend binaries"
go mod tidy
go mod download
go build -o bin/pinetmonitord ./cmd/pinetmonitord
go build -o bin/pinetmonitor ./cmd/pinetmonitor
