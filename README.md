# PiNetMonitor

PiNetMonitor – Simple internet usage monitoring for Raspberry Pi and Orange Pi.

PiNetMonitor is a lightweight open-source internet usage monitoring system for SBC devices such as Raspberry Pi and Orange Pi. It turns a Linux SBC into a gateway-based internet usage monitoring device that tracks traffic totals and produces daily and monthly reports from a local web dashboard.

PiNetMonitor is designed to be deployed as the real inline gateway for your home or lab network. For the supported production setup, use an SBC with **two Ethernet ports** or one built-in Ethernet port plus a **USB-to-Ethernet adapter**.

## Screenshots

Screenshot placeholders:

- `docs/screenshots/dashboard-overview.png`
- `docs/screenshots/daily-report.png`
- `docs/screenshots/monthly-report.png`

## Features

- Gateway-mode deployment for Debian, Raspberry Pi OS, Orange Pi, and Armbian style systems
- Inline whole-network deployment using two Ethernet interfaces so your existing home APs stay in place
- Daily and monthly internet usage reports backed by SQLite
- Lightweight web dashboard branded as `PiNetMonitor – Network Usage Dashboard`
- Simple operator CLI with `pinetmonitor status`, `pinetmonitor update`, `pinetmonitor restart`, and `pinetmonitor logs`
- Installer that uses a lighter SBC-friendly build path without pulling the full Node toolchain
- Production-minded architecture notes for HTTPS attribution, NAT, and low-resource ARM deployment
- Update pipeline scaffold for GitHub-hosted releases

## Recommended SBCs

For the supported inline PiNetMonitor deployment, it is best to use a board with two Ethernet ports, or a board that can reliably use a USB-to-Ethernet adapter for the second port.

Recommended boards from lower-cost to higher-end:

1. Orange Pi Zero 3
2. Raspberry Pi Zero 2 W with a supported USB-to-Ethernet adapter
3. Raspberry Pi 4 Model B with a supported USB-to-Ethernet adapter
4. Raspberry Pi 5 with a supported USB-to-Ethernet adapter

If your goal is whole-network monitoring with your existing APs unchanged, PiNetMonitor is easiest to deploy on a board that already has two Ethernet ports or can stably run with one built-in Ethernet port plus one USB Ethernet adapter.

## Install

Run the following command on your Raspberry Pi or Orange Pi:

```bash
curl -fsSL https://raw.githubusercontent.com/agraja38/PiNetMonitor/main/scripts/install.sh | sudo bash
```

The installer will:

- verify Linux compatibility
- verify that the SBC has two Ethernet interfaces for inline gateway mode
- install required runtime packages
- build the backend
- build the frontend
- initialize the database
- enable IP forwarding
- configure NAT rules
- install and start `pinetmonitor.service`

## Installation Instructions

1. Prepare a Debian-compatible SBC with two network interfaces if you want full gateway mode.
2. Use a device with two Ethernet ports, or add a USB-to-Ethernet adapter before installation.
3. Connect the future WAN port to your upstream router or modem, and the future LAN port to the switch or AP side of your network.
4. Review `configs/pinetmonitor.env.example` and plan your WAN, LAN, and LAN CIDR settings.
5. Run the one-line installer command above as `root` or with `sudo`. On low-resource SBCs, PiNetMonitor avoids `apt install golang-go` and instead bootstraps Go directly, while only installing the small native packages needed for SQLite builds.
6. Open `http://<device-ip>:8080` after installation to reach the dashboard.
7. Use `pinetmonitor status` to confirm that the service is healthy.

## Hardware Requirement

PiNetMonitor supports one deployment model:

- `WAN Ethernet` -> upstream router, modem, or ONT
- `LAN Ethernet` -> downstream switch or existing home AP infrastructure

PiNetMonitor does not target “sidecar” host-only monitoring or a separate Wi-Fi-only AP mode as the primary deployment path. If your goal is to monitor the whole network while keeping your existing APs, the Orange Pi or Raspberry Pi must be placed inline with two Ethernet interfaces.

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

1. Install Go 1.22+ and SQLite.
2. Clone the repository.
3. Copy `configs/pinetmonitor.env.example` to your working environment and adjust interface names.
4. Build the frontend:

```bash
./scripts/build-frontend.sh
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

PiNetMonitor can update from the public repository without a token. For higher API limits or private forks, provide a token at runtime instead of hard-coding it in the repository:

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
