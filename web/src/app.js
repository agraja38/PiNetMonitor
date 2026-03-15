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

const formatPercent = (value) => `${value.toFixed(1)}%`;

const renderRows = (targetId, rows, firstKey) => {
  const target = document.getElementById(targetId);
  if (!rows?.length) {
    target.innerHTML = `<tr><td colspan="4">No data collected yet.</td></tr>`;
    return;
  }

  target.innerHTML = rows
    .map(
      (row) => `
        <tr>
          <td>${row[firstKey]}</td>
          <td>${formatBytes(row.rx_bytes)}</td>
          <td>${formatBytes(row.tx_bytes)}</td>
          <td>${formatBytes(row.total_bytes)}</td>
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

const polarToCartesian = (cx, cy, radius, angleInDegrees) => {
  const angleInRadians = ((angleInDegrees - 90) * Math.PI) / 180.0;
  return {
    x: cx + radius * Math.cos(angleInRadians),
    y: cy + radius * Math.sin(angleInRadians)
  };
};

const describeArc = (x, y, radius, startAngle, endAngle) => {
  const start = polarToCartesian(x, y, radius, endAngle);
  const end = polarToCartesian(x, y, radius, startAngle);
  const largeArcFlag = endAngle - startAngle <= 180 ? "0" : "1";
  return `M ${x} ${y} L ${start.x} ${start.y} A ${radius} ${radius} 0 ${largeArcFlag} 0 ${end.x} ${end.y} Z`;
};

const renderServiceChart = (chartId, legendId, payload) => {
  const chart = document.getElementById(chartId);
  const legend = document.getElementById(legendId);

  if (!payload?.data_available || !payload.rows?.length) {
    chart.innerHTML = `
      <circle cx="60" cy="60" r="46" fill="#edf2e9"></circle>
      <circle cx="60" cy="60" r="24" fill="#ffffff"></circle>
      <text x="60" y="58" text-anchor="middle" class="chart-empty">Awaiting</text>
      <text x="60" y="70" text-anchor="middle" class="chart-empty">data</text>
    `;
    legend.innerHTML = `
      <div class="empty-card">
        Per-service attribution data is not available yet on this device.
      </div>
    `;
    return;
  }

  let startAngle = 0;
  chart.innerHTML = payload.rows
    .map((row) => {
      const angle = (row.percentage / 100) * 360;
      const path = describeArc(60, 60, 46, startAngle, startAngle + angle);
      startAngle += angle;
      return `<path d="${path}" fill="${row.color}"></path>`;
    })
    .join("") + `<circle cx="60" cy="60" r="24" fill="#ffffff"></circle>`;

  legend.innerHTML = payload.rows
    .map(
      (row) => `
        <div class="legend-row">
          <span class="swatch" style="background:${row.color}"></span>
          <div class="legend-copy">
            <strong>${row.service}</strong>
            <span>${formatBytes(row.usage_bytes)} · ${formatPercent(row.percentage)}</span>
          </div>
        </div>`
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
  const [statusRes, dailyRes, monthlyRes, dailyTopRes, monthlyTopRes] = await Promise.all([
    fetch("/api/v1/status"),
    fetch("/api/v1/reports/daily"),
    fetch("/api/v1/reports/monthly"),
    fetch("/api/v1/attribution/daily-top"),
    fetch("/api/v1/attribution/monthly-top")
  ]);

  const status = await statusRes.json();
  const daily = await dailyRes.json();
  const monthly = await monthlyRes.json();
  const dailyTop = await dailyTopRes.json();
  const monthlyTop = await monthlyTopRes.json();

  document.getElementById("service-status").textContent = `${status.name} ${status.version}`;
  document.getElementById("service-uptime").textContent = `Uptime ${status.uptime}`;
  renderFacts(status.gateway);
  renderSamples(status.samples);
  renderRows("daily-body", daily.rows, "bucket");
  renderRows("monthly-body", monthly.rows, "bucket");
  renderServiceChart("daily-services-chart", "daily-services-legend", dailyTop);
  renderServiceChart("monthly-services-chart", "monthly-services-legend", monthlyTop);
};

load().catch((error) => {
  document.getElementById("service-status").textContent = "Unable to load PiNetMonitor";
  document.getElementById("service-uptime").textContent = error.message;
});
