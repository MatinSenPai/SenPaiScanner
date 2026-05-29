const form = document.querySelector("#scanForm");
const startBtn = document.querySelector("#startBtn");
const cancelBtn = document.querySelector("#cancelBtn");
const copyBtn = document.querySelector("#copyBtn");
const refreshBtn = document.querySelector("#refreshBtn");
const statusEl = document.querySelector("#status");
const resultsBody = document.querySelector("#resultsBody");
const leaderboardBody = document.querySelector("#leaderboardBody");
const progressBar = document.querySelector("#progressBar");

let pollTimer = null;
let startedAt = null;
let lastState = null;

const fmt = new Intl.NumberFormat(undefined, { maximumFractionDigits: 1 });

async function api(path, options = {}) {
  const res = await fetch(path, {
    headers: { "Content-Type": "application/json" },
    ...options,
  });
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}

function payload() {
  const data = new FormData(form);
  return {
    count: Number(data.get("count")),
    concurrency: Number(data.get("concurrency")),
    timeout: String(data.get("timeout") || "3s"),
    tries: Number(data.get("tries")),
    port: Number(data.get("port")),
    mode: String(data.get("mode") || "http"),
    cidr: String(data.get("cidr") || ""),
    sni: String(data.get("sni") || ""),
    use_v4: data.has("use_v4"),
    use_v6: data.has("use_v6"),
    stealth: data.has("stealth"),
    fast: data.has("fast"),
    adaptive: data.has("adaptive"),
  };
}

function setMetric(id, value) {
  document.querySelector(id).textContent = value;
}

function row(cells) {
  const tr = document.createElement("tr");
  for (const cell of cells) {
    const td = document.createElement("td");
    if (cell.className) td.className = cell.className;
    td.textContent = cell.value;
    tr.appendChild(td);
  }
  return tr;
}

function renderResults(results = []) {
  resultsBody.replaceChildren();
  for (const r of results.slice(0, 120)) {
    resultsBody.appendChild(row([
      { value: r.ip, className: "ip" },
      { value: fmt.format(r.score), className: "score" },
      { value: `${fmt.format(r.loss_pct)}%` },
      { value: `${fmt.format(r.avg_ms)} ms` },
      { value: `${fmt.format(r.jitter_ms)} ms` },
      { value: `${fmt.format(r.throughput_kb)} KB/s` },
      { value: r.colo || "-" },
    ]));
  }
}

function renderLeaderboard(records = []) {
  leaderboardBody.replaceChildren();
  for (const r of records) {
    leaderboardBody.appendChild(row([
      { value: r.ip, className: "ip" },
      { value: fmt.format(r.score), className: "score" },
      { value: `${fmt.format(r.avg_ms)} ms` },
      { value: `${fmt.format((r.throughput_bps || 0) / 1024)} KB/s` },
      { value: r.colo || "-" },
      { value: String(r.seen_count || 0) },
    ]));
  }
}

function updateState(state) {
  if (!state || state.running === false) {
    statusEl.textContent = "Idle";
    return;
  }
  lastState = state;
  const stats = state.stats || {};
  setMetric("#tested", stats.tested || 0);
  setMetric("#healthy", stats.healthy || 0);
  setMetric("#failed", stats.failed || 0);

  const elapsed = Math.max(1, (Date.now() - startedAt) / 1000);
  setMetric("#rate", `${fmt.format((stats.tested || 0) / elapsed)}/s`);

  const target = Number(new FormData(form).get("count")) || 0;
  const pct = target > 0 ? Math.min(100, ((stats.tested || 0) / target) * 100) : 0;
  progressBar.style.width = `${pct}%`;
  renderResults(state.results || []);

  const suffix = state.error ? ` - ${state.error}` : "";
  statusEl.textContent = state.done ? `Done${suffix}` : `Running${suffix}`;
  startBtn.disabled = !state.done;
  cancelBtn.disabled = state.done;
}

async function poll() {
  try {
    updateState(await api("/api/state"));
    if (lastState?.done) {
      clearInterval(pollTimer);
      pollTimer = null;
      await loadLeaderboard();
    }
  } catch (err) {
    statusEl.textContent = err.message.trim();
  }
}

async function loadLeaderboard() {
  try {
    renderLeaderboard((await api("/api/leaderboard?limit=50")) || []);
  } catch (err) {
    statusEl.textContent = err.message.trim();
  }
}

startBtn.addEventListener("click", async () => {
  try {
    startBtn.disabled = true;
    cancelBtn.disabled = false;
    startedAt = Date.now();
    resultsBody.replaceChildren();
    progressBar.style.width = "0%";
    statusEl.textContent = "Starting";
    await api("/api/scan", { method: "POST", body: JSON.stringify(payload()) });
    resumePolling();
    await poll();
  } catch (err) {
    startBtn.disabled = false;
    statusEl.textContent = err.message.trim();
  }
});

cancelBtn.addEventListener("click", async () => {
  cancelBtn.disabled = true;
  await api("/api/cancel", { method: "POST", body: "{}" });
  await poll();
});

copyBtn.addEventListener("click", async () => {
  const ips = [...resultsBody.querySelectorAll("td.ip")].map((td) => td.textContent).join("\n");
  if (!ips) return;
  await navigator.clipboard.writeText(`${ips}\n`);
  statusEl.textContent = `Copied ${ips.split("\n").filter(Boolean).length} IPs`;
});

refreshBtn.addEventListener("click", loadLeaderboard);

function resumePolling() {
  if (pollTimer) clearInterval(pollTimer);
  pollTimer = setInterval(poll, 700);
}

async function loadPorts() {
  try {
    const ports = await api("/api/ports");
    const list = document.querySelector("#cfPorts");
    if (!list) return;
    list.replaceChildren();
    const add = (port, kind) => {
      const opt = document.createElement("option");
      opt.value = String(port);
      opt.label = `${port} (${kind})`;
      list.appendChild(opt);
    };
    (ports.https || []).forEach((p) => add(p, "HTTPS"));
    (ports.http || []).forEach((p) => add(p, "HTTP"));
  } catch {
    /* curated port hints are optional; ignore failures */
  }
}

async function boot() {
  cancelBtn.disabled = true;
  const version = await api("/api/version");
  document.querySelector("#version").textContent = `v${version.version}`;
  await loadPorts();
  await loadLeaderboard();

  // If a scan is already running (e.g. the page was reloaded mid-scan), pick the
  // clock back up from the server's start time and resume live polling instead
  // of showing a single frozen snapshot.
  const state = await api("/api/state");
  if (state && state.running !== false && !state.done) {
    startedAt = state.started ? new Date(state.started).getTime() : Date.now();
    resumePolling();
  }
  updateState(state);
}

boot();
