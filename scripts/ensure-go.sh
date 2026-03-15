#!/usr/bin/env bash
set -euo pipefail

GO_VERSION="${PINETMONITOR_GO_VERSION:-1.22.12}"
INSTALL_DIR="${PINETMONITOR_GO_INSTALL_DIR:-/usr/local}"

detect_go_arch() {
  case "$(uname -m)" in
    aarch64|arm64)
      echo "arm64"
      ;;
    armv7l|armv7*)
      echo "armv6l"
      ;;
    x86_64|amd64)
      echo "amd64"
      ;;
    *)
      echo ""
      ;;
  esac
}

GO_ARCH="$(detect_go_arch)"
if [ -z "$GO_ARCH" ]; then
  echo "[PiNetMonitor] Unsupported CPU architecture for Go bootstrap: $(uname -m)"
  exit 1
fi

if command -v go >/dev/null 2>&1; then
  CURRENT_VERSION="$(go version | awk '{print $3}')"
  if [ "$CURRENT_VERSION" = "go${GO_VERSION}" ]; then
    echo "[PiNetMonitor] Go ${GO_VERSION} already available"
    exit 0
  fi
fi

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

ARCHIVE="go${GO_VERSION}.linux-${GO_ARCH}.tar.gz"
URL="https://go.dev/dl/${ARCHIVE}"

echo "[PiNetMonitor] Downloading Go ${GO_VERSION} (${GO_ARCH})"
curl -fsSL "$URL" -o "${TMP_DIR}/${ARCHIVE}"
rm -rf "${INSTALL_DIR}/go"
tar -C "$INSTALL_DIR" -xzf "${TMP_DIR}/${ARCHIVE}"
echo "[PiNetMonitor] Installed Go ${GO_VERSION} to ${INSTALL_DIR}/go"
