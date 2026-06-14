const $ = (id) => document.getElementById(id);

let lastState = null;
let toastTimer = null;

const fmt = new Intl.NumberFormat("en-US");

function readNumber(id, fallback) {
  const n = Number($(id).value);
  return Number.isFinite(n) ? n : fallback;
}

function readForm() {
  const checkedPorts = [...document.querySelectorAll("#ports input:checked")]
    .map((el) => Number(el.value));

  return {
    source: document.querySelector("input[name='source']:checked").value,
    count: readNumber("count", 5000),
    workers: readNumber("workers", 50),
    timeout: $("timeout").value.trim() || "5s",
    ports: checkedPorts.filter((port) => port > 0),
    use_config_port: checkedPorts.includes(0),
    config_url: $("configUrl").value.trim(),
    top_n: readNumber("topN", 50),
    min_speed_mbps: readNumber("minSpeed", 0),
    speed_size_kb: readNumber("speedSize", 512),
    upload_test: $("uploadTest").checked,
    require_ws: $("requireWs").checked,
  };
}

async function request(path, options = {}) {
  const res = await fetch(path, {
    headers: { "Content-Type": "application/json" },
    ...options,
  });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    throw new Error(data.error || res.statusText);
  }
  return data;
}

async function startScan() {
  try {
    await request("/api/scan/start", {
      method: "POST",
      body: JSON.stringify(readForm()),
    });
    toast("Scan started");
    await refreshState();
  } catch (err) {
    toast(err.message);
  }
}

async function stopScan() {
  try {
    await request("/api/scan/stop", { method: "POST", body: "{}" });
    toast("Stopping scan");
  } catch (err) {
    toast(err.message);
  }
}

async function saveResults() {
  try {
    const data = await request("/api/results/save", { method: "POST", body: "{}" });
    toast(`Saved ${data.count} endpoints to ${data.path}`);
    await refreshState();
  } catch (err) {
    toast(err.message);
  }
}

async function copyResults() {
  const endpoints = (lastState?.working || []).map((row) => row.endpoint);
  if (!endpoints.length) {
    toast("No working endpoints to copy");
    return;
  }
  try {
    await navigator.clipboard.writeText(endpoints.join("\n") + "\n");
    toast(`Copied ${endpoints.length} endpoints`);
  } catch (err) {
    toast(`Clipboard failed: ${err.message}`);
  }
}

async function quitApp() {
  try {
    await request("/api/app/quit", { method: "POST", body: "{}" });
    toast("Closing app");
  } catch (err) {
    toast(err.message);
  }
}

async function refreshState() {
  try {
    const state = await request("/api/scan/state");
    lastState = state;
    render(state);
  } catch (err) {
    $("statusText").textContent = err.message;
  }
}

async function loadMeta() {
  try {
    const meta = await request("/api/meta");
    document.title = meta.name || "MAGIK";
    $("version").textContent = meta.version || "dev";
  } catch {
    document.title = "MAGIK";
    $("version").textContent = "dev";
  }
}

function render(state) {
  const running = Boolean(state.running);
  const working = state.working || [];
  const results = state.results || [];
  const phase1 = state.phase1 || {};
  const phase2 = state.phase2 || {};

  $("startBtn").disabled = running;
  $("stopBtn").disabled = !running;
  $("copyBtn").disabled = !working.length;
  $("saveBtn").disabled = !working.length;

  $("statusText").textContent = state.error || state.status || "Ready";
  $("phaseLabel").textContent = state.phase || "idle";
  $("resultCount").textContent = `${fmt.format(results.length)} rows`;
  $("workingCount").textContent = fmt.format(working.length);
  $("p1Done").textContent = `${fmt.format(phase1.done || 0)} / ${fmt.format(phase1.total || 0)}`;
  $("p1Healthy").textContent = fmt.format(phase1.healthy || 0);
  $("p2Done").textContent = `${fmt.format(phase2.done || 0)} / ${fmt.format(phase2.total || 0)}`;
  $("elapsed").textContent = formatElapsed(state.elapsed_seconds || 0);
  $("saveState").textContent = state.save_path ? "saved" : "not saved";

  const dot = $("phaseDot");
  dot.className = "dot";
  if (state.phase === "error") dot.classList.add("error");
  else if (running) dot.classList.add("running");
  else if (state.done) dot.classList.add("done");
  else dot.classList.add("idle");

  const activeStats = state.phase === "phase2" ? phase2 : phase1;
  const pct = activeStats.total > 0 ? Math.min(100, (activeStats.done / activeStats.total) * 100) : 0;
  $("progressBar").style.width = `${pct}%`;

  renderWorking(working);
  renderResults(results);
}

function renderWorking(rows) {
  const box = $("workingList");
  if (!rows.length) {
    box.className = "working-list empty";
    box.textContent = "No working endpoints yet";
    return;
  }

  box.className = "working-list";
  box.innerHTML = rows.slice(0, 120).map((row) => {
    const metric = row.speed_mbps > 0
      ? `${row.speed_mbps.toFixed(1)} Mbps`
      : `${row.avg_ms.toFixed(1)} ms`;
    return `
      <div class="endpoint-row">
        <code title="${escapeHtml(row.endpoint)}">${escapeHtml(row.endpoint)}</code>
        <span>${escapeHtml(metric)}</span>
      </div>
    `;
  }).join("");
}

function renderResults(rows) {
  const body = $("resultsBody");
  if (!rows.length) {
    body.innerHTML = `<tr><td colspan="7" class="empty-cell">Ready</td></tr>`;
    return;
  }

  body.innerHTML = rows.slice(-180).reverse().map((row) => {
    const cls = row.is_working ? "ok" : (row.status === "failed" ? "bad" : "warn");
    const speed = row.speed_mbps > 0 ? `${row.speed_mbps.toFixed(1)} Mbps` : "-";
    const avg = row.latency_ms > 0
      ? `${row.latency_ms.toFixed(1)} ms`
      : (row.avg_ms > 0 ? `${row.avg_ms.toFixed(1)} ms` : "-");
    const loss = Number.isFinite(row.loss_pct) && row.phase === "phase1" ? `${row.loss_pct.toFixed(1)}%` : "-";
    const status = row.error ? `${row.status}: ${row.error}` : row.status;
    return `
      <tr>
        <td class="muted">${escapeHtml(row.phase)}</td>
        <td>${escapeHtml(row.endpoint)}</td>
        <td>${escapeHtml(row.colo || "-")}</td>
        <td>${escapeHtml(loss)}</td>
        <td>${escapeHtml(avg)}</td>
        <td>${escapeHtml(speed)}</td>
        <td class="${cls}" title="${escapeHtml(status)}">${escapeHtml(row.status)}</td>
      </tr>
    `;
  }).join("");
}

function formatElapsed(seconds) {
  const s = Math.max(0, Math.floor(seconds));
  const m = Math.floor(s / 60);
  const r = s % 60;
  return `${String(m).padStart(2, "0")}:${String(r).padStart(2, "0")}`;
}

function escapeHtml(value) {
  return String(value)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#039;");
}

function toast(message) {
  const el = $("toast");
  el.textContent = message;
  el.hidden = false;
  clearTimeout(toastTimer);
  toastTimer = setTimeout(() => {
    el.hidden = true;
  }, 3600);
}

function applyTheme(dark) {
  document.documentElement.dataset.theme = dark ? "dark" : "light";
  localStorage.setItem("magik-theme", dark ? "dark" : "light");
  $("themeToggle").checked = dark;
}

function bootTheme() {
  const saved = localStorage.getItem("magik-theme") || localStorage.getItem("senpai-theme");
  applyTheme(saved !== "light");
}

$("startBtn").addEventListener("click", startScan);
$("stopBtn").addEventListener("click", stopScan);
$("saveBtn").addEventListener("click", saveResults);
$("copyBtn").addEventListener("click", copyResults);
$("quitBtn").addEventListener("click", quitApp);
$("themeToggle").addEventListener("change", (event) => applyTheme(event.target.checked));

bootTheme();
loadMeta();
refreshState();
setInterval(refreshState, 700);
