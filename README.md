# PiNetMonitor

PiNetMonitor – Simple internet usage monitoring for Raspberry Pi and Orange Pi.

PiNetMonitor is a lightweight open-source internet usage monitoring system for SBC devices such as Raspberry Pi and Orange Pi. It turns a Linux SBC into a gateway-based internet usage monitoring device that tracks traffic totals and produces daily and monthly reports from a local web dashboard.

## Screenshots

Screenshot placeholders:

- `docs/screenshots/dashboard-overview.png`
- `docs/screenshots/daily-report.png`
- `docs/screenshots/monthly-report.png`

## Features

- Gateway-mode deployment for Debian, Raspberry Pi OS, Orange Pi, and Armbian style systems
- Daily and monthly internet usage reports backed by SQLite
- Lightweight web dashboard branded as `PiNetMonitor – Network Usage Dashboard`
- Simple operator CLI with `pinetmonitor status`, `pinetmonitor update`, `pinetmonitor restart`, and `pinetmonitor logs`
- Installer that sets up dependencies, networking, NAT, database, frontend assets, and systemd services
- Production-minded architecture notes for HTTPS attribution, NAT, and low-resource ARM deployment
- Update pipeline scaffold for GitHub-hosted releases

## Install

Run the following command on your Raspberry Pi or Orange Pi:

```bash
curl -fsSL https://raw.githubusercontent.com/agraja38/PiNetMonitor/main/scripts/install.sh | bash
```

The installer will:

- verify Linux compatibility
- install required packages
- build the backend
- build the frontend
- initialize the database
- enable IP forwarding
- configure NAT rules
- install and start `pinetmonitor.service`

## Installation Instructions

1. Prepare a Debian-compatible SBC with two network interfaces if you want full gateway mode.
2. Review `configs/pinetmonitor.env.example` and plan your WAN, LAN, and LAN CIDR settings.
3. Run the one-line installer command above as `root` or with `sudo`.
4. Open `http://<device-ip>:8080` after installation to reach the dashboard.
5. Use `pinetmonitor status` to confirm that the service is healthy.

## Architecture Overview

PiNetMonitor is built around a small Go daemon, a static web UI, SQLite storage, and system-managed routing on Linux.

- `cmd/pinetmonitord`: backend daemon and HTTP API
- `cmd/pinetmonitor`: CLI tool for operators
- `internal/collector`: interface sampler with room for future flow attribution backends
- `internal/store`: SQLite-backed reporting store
- `scripts/configure-network.sh`: gateway routing and NAT setup
- `packaging/systemd/pinetmonitor.service`: long-running service definition

More detail is available in [docs/architecture.md](docs/architecture.md).

## Development Setup

1. Install Go 1.22+, Node.js 20+, npm, SQLite, and build tools.
2. Clone the repository.
3. Copy `configs/pinetmonitor.env.example` to your working environment and adjust interface names.
4. Build the frontend:

```bash
cd web
npm run build
```

5. Build the backend:

```bash
go build -o bin/pinetmonitord ./cmd/pinetmonitord
go build -o bin/pinetmonitor ./cmd/pinetmonitor
```

6. Start the daemon with the environment variables exported for your test environment.

## Update Instructions

For local updates on a deployed box:

```bash
sudo pinetmonitor update
```

PiNetMonitor expects a GitHub token to be provided at runtime instead of hard-coded in the repository:

- `GITHUB_USERNAME=agraja38`
- `GITHUB_REPOSITORY=agraja38/PiNetMonitor`
- `GITHUB_ACCESS_TOKEN` should be exported in the shell or stored securely in your environment management system before running updates

## Contribution Guide

1. Open an issue describing the bug, hardware target, or proposed feature.
2. Keep SBC resource usage in mind when proposing new services or dependencies.
3. Prefer reversible networking changes and explicit migration steps.
4. Test gateway mode on real Debian-based ARM hardware when your change touches routing or update behavior.

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE).

## Roadmap

- Add conntrack-backed flow capture and DNS correlation tables
- Add device-level attribution with confidence scoring for HTTPS destinations
- Add signed release manifests and rollback-safe OTA updates
- Add per-device quotas and alerts
- Add exportable CSV and JSON reports
