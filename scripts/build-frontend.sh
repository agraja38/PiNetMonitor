#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SRC_DIR="${ROOT_DIR}/web/src"
DIST_DIR="${ROOT_DIR}/web/dist"

echo "[PiNetMonitor] Building frontend"
rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"
cp "$SRC_DIR/index.html" "$DIST_DIR/index.html"
cp "$SRC_DIR/styles.css" "$DIST_DIR/styles.css"
cp "$SRC_DIR/app.js" "$DIST_DIR/app.js"
