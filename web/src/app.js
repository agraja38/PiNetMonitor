const formatBytes = (value) => {
  if (typeof value !== "number") return "n/a";
  const units = ["B", "KB", "MB", "GB", "TB"];
  let size = value;
  let index = 0;
  while (size >= 1024 && index < units.length - 1) {
    size /= 1024;
    index += 1;
  }
  return `${size.toFixed(index === 0 ? 0 : 2)} ${units[index]}`;
};

const renderRows = (targetId, rows, firstKey) => {
  const target = document.getElementById(targetId);
  if (!rows?.length) {
    target.innerHTML = `<tr><td colspan="5">No data collected yet.</td></tr>`;
    return;
  }

  target.innerHTML = rows
    .map(
      (row) => `
        <tr>
          <td>${row[firstKey]}</td>
          <td>${row.interface}</td>
          <td>${formatBytes(row.total_bytes)}</td>
          <td>${formatBytes(row.rx_bytes)}</td>
          <td>${formatBytes(row.tx_bytes)}</td>
        </tr>`
    )
    .join("");
};

const renderSamples = (samples) => {
  const target = document.getElementById("samples-body");
  if (!samples?.length) {
    target.innerHTML = `<tr><td colspan="4">Waiting for interface samples.</td></tr>`;
    return;
  }

  target.innerHTML = samples
    .map(
      (sample) => `
        <tr>
          <td>${sample.interface}</td>
          <td>${formatBytes(sample.rx_bytes)}</td>
          <td>${formatBytes(sample.tx_bytes)}</td>
          <td>${new Date(sample.timestamp).toLocaleString()}</td>
        </tr>`
    )
    .join("");
};

const renderFacts = (gateway) => {
  const facts = [
    ["Gateway Mode", gateway.mode ? "Enabled" : "Disabled"],
    ["NAT", gateway.nat_enabled ? "Enabled" : "Disabled"],
    ["WAN Interface", gateway.wan_interface],
    ["LAN Interface", gateway.lan_interface],
    ["LAN CIDR", gateway.lan_cidr],
    ["HTTPS Attribution", gateway.https_attribution ? "Enabled" : "Disabled"],
    ["Strategy", gateway.attribution_strategy],
    ["Collector", gateway.capture_backend]
  ];

  document.getElementById("gateway-facts").innerHTML = facts
    .map(
      ([label, value]) => `
        <div>
          <dt>${label}</dt>
          <dd>${value}</dd>
        </div>`
    )
    .join("");
};

const load = async () => {
  const [statusRes, dailyRes, monthlyRes] = await Promise.all([
    fetch("/api/v1/status"),
    fetch("/api/v1/reports/daily"),
    fetch("/api/v1/reports/monthly")
  ]);

  const status = await statusRes.json();
  const daily = await dailyRes.json();
  const monthly = await monthlyRes.json();

  document.getElementById("service-status").textContent = `${status.name} ${status.version}`;
  document.getElementById("service-uptime").textContent = `Uptime ${status.uptime}`;
  renderFacts(status.gateway);
  renderSamples(status.samples);
  renderRows("daily-body", daily.rows, "bucket");
  renderRows("monthly-body", monthly.rows, "bucket");
};

load().catch((error) => {
  document.getElementById("service-status").textContent = "Unable to load PiNetMonitor";
  document.getElementById("service-uptime").textContent = error.message;
});
