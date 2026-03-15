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
let updatePoller;
let settingsBound = false;

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

const fillSettings = (settings) => {
  document.getElementById("wan-interface").value = settings.wan_interface ?? "";
  document.getElementById("lan-interface").value = settings.lan_interface ?? "";
  document.getElementById("lan-cidr").value = settings.lan_cidr ?? "";
  document.getElementById("sample-interval").value = settings.sample_interval ?? "30s";
  document.getElementById("github-repository").value = settings.github_repository ?? "";
  document.getElementById("github-access-token").value = "";
  document.getElementById("gateway-mode").checked = Boolean(settings.gateway_mode);
  document.getElementById("enable-nat").checked = Boolean(settings.enable_nat);
  document.getElementById("enable-https-attribution").checked = Boolean(settings.enable_https_attribution);
};

const bindNavigation = () => {
  const links = Array.from(document.querySelectorAll(".nav-link"));
  const panels = Array.from(document.querySelectorAll("[data-view-panel]"));

  const setView = (view) => {
    links.forEach((link) => {
      link.classList.toggle("active", link.dataset.view === view);
    });
    panels.forEach((panel) => {
      panel.classList.toggle("active", panel.dataset.viewPanel === view);
    });
  };

  links.forEach((link) => {
    link.addEventListener("click", () => setView(link.dataset.view));
  });
};

const renderUpdateStatus = (status) => {
  const banner = document.getElementById("update-banner");
  const version = document.getElementById("update-version");
  const state = document.getElementById("update-state");
  const log = document.getElementById("update-log");

  if (status.update_available) {
    banner.textContent = `Update available: v${status.latest_version}`;
    banner.className = "update-banner available";
  } else if (status.check_error) {
    banner.textContent = "Update check failed";
    banner.className = "update-banner warning";
  } else {
    banner.textContent = "PiNetMonitor is up to date";
    banner.className = "update-banner";
  }

  version.textContent = `Current v${status.current_version} · Latest v${status.latest_version || status.current_version}`;
  state.textContent = status.running ? "Update in progress" : `Last exit code: ${status.last_exit_code}`;
  log.textContent = status.log?.trim() || "Waiting for update activity...";
  document.getElementById("start-update").disabled = Boolean(status.running);
};

const fetchUpdateStatus = async () => {
  const response = await fetch("/api/v1/update/status");
  const status = await response.json();
  renderUpdateStatus(status);
};

const startUpdatePolling = () => {
  window.clearInterval(updatePoller);
  updatePoller = window.setInterval(() => {
    fetchUpdateStatus().catch(() => {});
  }, 5000);
};

const bindSettings = () => {
  if (settingsBound) {
    return;
  }
  settingsBound = true;

  document.getElementById("settings-form").addEventListener("submit", async (event) => {
    event.preventDefault();
    const message = document.getElementById("settings-message");
    message.textContent = "Saving...";

    const payload = {
      wan_interface: document.getElementById("wan-interface").value,
      lan_interface: document.getElementById("lan-interface").value,
      lan_cidr: document.getElementById("lan-cidr").value,
      sample_interval: document.getElementById("sample-interval").value,
      github_repository: document.getElementById("github-repository").value,
      github_access_token: document.getElementById("github-access-token").value,
      gateway_mode: document.getElementById("gateway-mode").checked,
      enable_nat: document.getElementById("enable-nat").checked,
      enable_https_attribution: document.getElementById("enable-https-attribution").checked,
      http_addr: "0.0.0.0:8080"
    };

    const response = await fetch("/api/v1/settings", {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload)
    });
    const data = await response.json();
    if (!response.ok) {
      message.textContent = data.error || "Unable to save settings";
      return;
    }
    message.textContent = "Saved. Restart PiNetMonitor if you changed listener settings.";
  });

  document.getElementById("refresh-updates").addEventListener("click", () => {
    fetchUpdateStatus().catch(() => {});
  });

  document.getElementById("start-update").addEventListener("click", async () => {
    const response = await fetch("/api/v1/update/start", { method: "POST" });
    const data = await response.json();
    if (!response.ok) {
      document.getElementById("update-log").textContent = data.error || "Unable to start update";
      return;
    }
    await fetchUpdateStatus();
  });
};

const load = async () => {
  const [statusRes, dailyRes, monthlyRes, dailyTopRes, monthlyTopRes, settingsRes, updateRes] = await Promise.all([
    fetch("/api/v1/status"),
    fetch("/api/v1/reports/daily"),
    fetch("/api/v1/reports/monthly"),
    fetch("/api/v1/attribution/daily-top"),
    fetch("/api/v1/attribution/monthly-top"),
    fetch("/api/v1/settings"),
    fetch("/api/v1/update/status")
  ]);

  const status = await statusRes.json();
  const daily = await dailyRes.json();
  const monthly = await monthlyRes.json();
  const dailyTop = await dailyTopRes.json();
  const monthlyTop = await monthlyTopRes.json();
  const settings = await settingsRes.json();
  const update = await updateRes.json();

  document.getElementById("service-status").textContent = `${status.name} ${status.version}`;
  document.getElementById("service-uptime").textContent = `Uptime ${status.uptime}`;
  renderFacts(status.gateway);
  renderSamples(status.samples);
  renderRows("daily-body", daily.rows, "bucket");
  renderRows("monthly-body", monthly.rows, "bucket");
  renderServiceChart("daily-services-chart", "daily-services-legend", dailyTop);
  renderServiceChart("monthly-services-chart", "monthly-services-legend", monthlyTop);
  fillSettings(settings);
  renderUpdateStatus(update);
  bindNavigation();
  bindSettings();
  startUpdatePolling();
};

load().catch((error) => {
  document.getElementById("service-status").textContent = "Unable to load PiNetMonitor";
  document.getElementById("service-uptime").textContent = error.message;
});
