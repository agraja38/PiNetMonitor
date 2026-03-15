#!/usr/bin/env bash
set -euo pipefail

LAN_IFACE="${PINETMONITOR_LAN_IFACE:-eth1}"
WAN_IFACE="${PINETMONITOR_WAN_IFACE:-eth0}"
LAN_CIDR="${PINETMONITOR_LAN_CIDR:-192.168.50.0/24}"
NFT_FILE="/etc/nftables.d/pinetmonitor.nft"

echo "[PiNetMonitor] Enabling IPv4 forwarding"
install -d -m 0755 /etc/sysctl.d
cat >/etc/sysctl.d/99-pinetmonitor.conf <<EOF
net.ipv4.ip_forward=1
net.ipv6.conf.all.forwarding=1
EOF
sysctl --system >/dev/null

echo "[PiNetMonitor] Installing nftables rules for ${LAN_IFACE} -> ${WAN_IFACE}"
install -d -m 0755 /etc/nftables.d
cat >"$NFT_FILE" <<EOF
table inet pinetmonitor_filter {
  chain forward {
    type filter hook forward priority 0;
    policy drop;
    ct state established,related accept
    iifname "${LAN_IFACE}" oifname "${WAN_IFACE}" accept
    iifname "${WAN_IFACE}" oifname "${LAN_IFACE}" ct state established,related accept
  }
}

table ip pinetmonitor_nat {
  chain postrouting {
    type nat hook postrouting priority 100;
    oifname "${WAN_IFACE}" ip saddr ${LAN_CIDR} masquerade
  }
}
EOF

if ! grep -q 'include "/etc/nftables.d/*.nft"' /etc/nftables.conf 2>/dev/null; then
  cat >/etc/nftables.conf <<'EOF'
#!/usr/sbin/nft -f
flush ruleset
include "/etc/nftables.d/*.nft"
EOF
fi

nft -f /etc/nftables.conf
systemctl enable nftables >/dev/null 2>&1 || true
systemctl restart nftables

echo "[PiNetMonitor] Network forwarding and NAT configured"
