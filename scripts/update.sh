#!/usr/bin/env bash
set -euo pipefail

INSTALL_ROOT="${INSTALL_ROOT:-/opt/pinetmonitor}"
REPO="${PINETMONITOR_GITHUB_REPOSITORY:-agraja38/PiNetMonitor}"
BRANCH="${PINETMONITOR_UPDATE_BRANCH:-main}"
TOKEN="${GITHUB_ACCESS_TOKEN:-}"

echo "[PiNetMonitor] Starting update"
if [ -z "$TOKEN" ]; then
  echo "[PiNetMonitor] GITHUB_ACCESS_TOKEN is required for authenticated update checks"
  exit 1
fi

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

ARCHIVE_URL="https://api.github.com/repos/${REPO}/tarball/${BRANCH}"
ARCHIVE_PATH="${TMP_DIR}/pinetmonitor.tar.gz"

echo "[PiNetMonitor] Downloading latest source from ${REPO}"
curl -fsSL \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Accept: application/vnd.github+json" \
  "$ARCHIVE_URL" \
  -o "$ARCHIVE_PATH"

tar -xzf "$ARCHIVE_PATH" -C "$TMP_DIR"
SRC_DIR="$(find "$TMP_DIR" -mindepth 1 -maxdepth 1 -type d | head -n 1)"

echo "[PiNetMonitor] Rebuilding release"
cd "$SRC_DIR"
./scripts/build-frontend.sh
./scripts/build-backend.sh

echo "[PiNetMonitor] Deploying new version"
install -d -m 0755 "${INSTALL_ROOT}/bin" "${INSTALL_ROOT}/web"
install -m 0755 bin/pinetmonitord "${INSTALL_ROOT}/bin/pinetmonitord"
install -m 0755 bin/pinetmonitor "${INSTALL_ROOT}/bin/pinetmonitor"
rm -rf "${INSTALL_ROOT}/web/dist"
cp -R web/dist "${INSTALL_ROOT}/web/dist"

systemctl restart pinetmonitor.service
echo "[PiNetMonitor] Update completed"
