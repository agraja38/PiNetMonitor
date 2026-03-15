#!/usr/bin/env bash
set -euo pipefail

PROJECT_NAME="PiNetMonitor"
INSTALL_ROOT="/opt/pinetmonitor"
SERVICE_NAME="pinetmonitor.service"
ENV_FILE="/etc/pinetmonitor/pinetmonitor.env"
REPO_URL="https://github.com/agraja38/PiNetMonitor.git"

log() {
  echo "[${PROJECT_NAME}] $1"
}

require_root() {
  if [ "${EUID}" -ne 0 ]; then
    log "Please run the installer as root or with sudo."
    exit 1
  fi
}

detect_platform() {
  if [ ! -f /etc/os-release ]; then
    log "Unsupported Linux distribution: missing /etc/os-release"
    exit 1
  fi

  . /etc/os-release
  case "${ID:-}" in
    debian|raspbian|ubuntu|armbian)
      log "Detected supported distribution: ${PRETTY_NAME:-$ID}"
      ;;
    *)
      log "Unsupported distribution: ${PRETTY_NAME:-$ID}"
      exit 1
      ;;
  esac

  DEVICE_MODEL="$(tr -d '\0' </proc/device-tree/model 2>/dev/null || true)"
  if [[ "$DEVICE_MODEL" == *"Raspberry Pi"* ]] || [[ "$DEVICE_MODEL" == *"Orange Pi"* ]]; then
    log "Detected SBC platform: ${DEVICE_MODEL}"
  else
    log "Proceeding on Debian-compatible Linux host: ${DEVICE_MODEL:-generic device}"
  fi
}

install_packages() {
  log "Installing system packages"
  apt-get update
  DEBIAN_FRONTEND=noninteractive apt-get install -y \
    git curl ca-certificates build-essential pkg-config \
    sqlite3 libsqlite3-dev nftables iproute2 nodejs npm golang-go
}

fetch_source() {
  log "Preparing installation directory"
  install -d -m 0755 "$INSTALL_ROOT" /etc/pinetmonitor /var/lib/pinetmonitor /var/log/pinetmonitor
  if [ ! -d "${INSTALL_ROOT}/.git" ]; then
    git clone "$REPO_URL" "$INSTALL_ROOT"
  else
    git -C "$INSTALL_ROOT" fetch --all --tags
    git -C "$INSTALL_ROOT" reset --hard origin/main
  fi
}

write_env() {
  log "Writing PiNetMonitor environment configuration"
  if [ ! -f "$ENV_FILE" ]; then
    install -m 0644 "${INSTALL_ROOT}/configs/pinetmonitor.env.example" "$ENV_FILE"
  fi
}

build_components() {
  log "Building backend"
  (cd "$INSTALL_ROOT" && ./scripts/build-backend.sh)
  log "Building frontend"
  (cd "$INSTALL_ROOT" && ./scripts/build-frontend.sh)
}

init_database() {
  log "Initializing database"
  export PINETMONITOR_DB_PATH="/var/lib/pinetmonitor/pinetmonitor.db"
  (cd "$INSTALL_ROOT" && ./scripts/init-db.sh)
}

configure_network() {
  log "Configuring routing, forwarding, and NAT"
  set -a
  . "$ENV_FILE"
  set +a
  if [ "${PINETMONITOR_GATEWAY_MODE:-true}" = "true" ] && [ "${PINETMONITOR_ENABLE_NAT:-true}" = "true" ]; then
    (cd "$INSTALL_ROOT" && ./scripts/configure-network.sh)
  else
    log "Gateway mode disabled in ${ENV_FILE}; skipping NAT setup"
  fi
}

install_service_units() {
  log "Installing systemd service units"
  install -m 0644 "${INSTALL_ROOT}/packaging/systemd/pinetmonitor.service" /etc/systemd/system/pinetmonitor.service
  install -m 0644 "${INSTALL_ROOT}/packaging/systemd/pinetmonitor-updater.service" /etc/systemd/system/pinetmonitor-updater.service
  install -m 0644 "${INSTALL_ROOT}/packaging/systemd/pinetmonitor-updater.timer" /etc/systemd/system/pinetmonitor-updater.timer
  systemctl daemon-reload
  systemctl enable "$SERVICE_NAME"
}

install_cli() {
  log "Installing pinetmonitor CLI"
  install -m 0755 "${INSTALL_ROOT}/bin/pinetmonitor" /usr/local/bin/pinetmonitor
}

start_services() {
  log "Starting PiNetMonitor service"
  systemctl restart "$SERVICE_NAME"
  log "PiNetMonitor is now running"
}

main() {
  require_root
  detect_platform
  install_packages
  fetch_source
  write_env
  build_components
  init_database
  configure_network
  install_service_units
  install_cli
  start_services
}

main "$@"
