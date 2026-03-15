# PiNetMonitor Architecture

PiNetMonitor is built as a gateway-first monitoring system for SBCs such as Raspberry Pi and Orange Pi. The service is designed to be useful on small ARM boards without hiding the tradeoffs that come with modern encrypted traffic.

PiNetMonitor’s supported deployment target is an inline gateway with two Ethernet interfaces. One interface faces upstream toward the existing internet connection, and the other faces downstream toward the home switch and access points.

## Core Components

- `pinetmonitord`: the daemon that samples interfaces, stores counters in SQLite, and serves the local web UI and API.
- `pinetmonitor`: the operator CLI for service status, restarts, logs, and updates.
- `nftables` and `sysctl`: the routing and NAT control plane for turning the SBC into a monitored gateway.
- SQLite in WAL mode: durable local storage with low operational overhead on Debian-based systems.
- Static web UI: lightweight dashboard that can be built on-device without pulling a large frontend toolchain.

## Traffic Attribution Strategy For HTTPS

PiNetMonitor does not pretend that encrypted traffic can always be attributed to exact URLs. The attribution pipeline is intentionally layered:

1. Flow identity from source IP, destination IP, port, protocol, and time.
2. DNS correlation from resolver responses observed near flow establishment, bounded by TTL and time windows.
3. TLS metadata such as SNI when a client exposes it.
4. QUIC and TLS handshake metadata when available.
5. Confidence scoring on every inferred mapping.

This keeps the system honest. For many workloads PiNetMonitor can say that a device used `youtube.com` or `api.openai.com` with medium or high confidence, but it will not claim visibility into exact HTTPS paths or content unless a future deployment explicitly adds an intercepting proxy.

## Gateway and NAT Design

PiNetMonitor assumes two Ethernet interfaces in gateway mode:

- WAN: upstream internet Ethernet interface such as `eth0` or `end0`
- LAN: downstream Ethernet interface such as `eth1` or a USB NIC

This design keeps existing home APs in place and routes all user traffic through PiNetMonitor before it reaches the internet. That is the intended production architecture for whole-network monitoring.

`nftables` is used instead of legacy `iptables` because it is the forward-looking packet filter on modern Debian-based systems and is easier to manage idempotently. The installer writes a dedicated ruleset include file instead of mutating unrelated firewall state inline.

## Storage Model

The first production milestone stores interface byte counters in SQLite so PiNetMonitor can provide daily and monthly reporting with very little CPU overhead. This is intentionally conservative for ARM SBCs:

- no daemon-local in-memory time series database
- no message bus requirement
- low-write WAL-backed SQLite
- straightforward backup and recovery

Flow-level tables, DNS event tables, and attribution confidence tables are planned as follow-on migrations once the capture pipeline is added.

## OTA Update Model

The initial update path downloads a source archive from GitHub, rebuilds the binaries locally, and restarts the systemd service. In a production deployment this should evolve toward:

- signed release manifests
- detached artifact verification
- staged update validation
- rollback on failed health checks

The repository already isolates update logic in `scripts/update.sh` and `pinetmonitor-updater.service` so the delivery mechanism can be improved without changing the daemon.
