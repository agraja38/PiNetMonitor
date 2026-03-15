#!/usr/bin/env bash
set -euo pipefail

DB_PATH="${PINETMONITOR_DB_PATH:-/var/lib/pinetmonitor/pinetmonitor.db}"

mkdir -p "$(dirname "$DB_PATH")"
if [ ! -f "$DB_PATH" ]; then
  echo "[PiNetMonitor] Initializing SQLite database at $DB_PATH"
  sqlite3 "$DB_PATH" < "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/migrations/001_init.sql"
else
  echo "[PiNetMonitor] Database already exists at $DB_PATH"
fi
