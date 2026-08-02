/* ===== SenPai Scanner — frontend logic (Wails v2) ===== */
"use strict";

const $ = (id) => document.getElementById(id);

// --- shared state -----------------------------------------------------------
const state = {
  count: 5000,
  workers: 50,
  mode: "http",
  timeoutMs: 5000,
  port: 443,
  sni: "speed.cloudflare.com",
  coloFilter: "",
  speedTest: true,
  requireWS: true,
  scanning: false,
  healthy: [],        // ScanResult[]
  liveResults: [],    // ScanResult[]
  validation: [],     // ValidationOutcome[]
  exported: null,     // ExportBundle
};

// --- toast ------------------------------------------------------------------
let toastTimer = null;
function toast(msg, ms = 3200) {
  const el = $("toast");
  el.textContent = msg;
  el.classList.add("show");
  clearTimeout(toastTimer);
  toastTimer = setTimeout(() => el.classList.remove("show"), ms);
}

// --- tab switching ------------------------------------------------------------
document.querySelectorAll(".nav-item").forEach((btn) => {
  btn.addEventListener("click", () => {
    document.querySelectorAll(".nav-item").forEach((b) => b.classList.remove("active"));
    btn.classList.add("active");
    document.querySelectorAll(".tab-pane").forEach((p) => p.classList.add("hidden"));
    $("tab-" + btn.dataset.tab).classList.remove("hidden");
  });
});

// --- segmented controls -------------------------------------------------------
function wireSeg(id, onChange) {
  const seg = $(id);
  seg.addEventListener("click", (e) => {
    const btn = e.target.closest("button");
    if (!btn) return;
    seg.querySelectorAll("button").forEach((b) => b.classList.remove("on"));
    btn.classList.add("on");
    onChange(Number(btn.dataset.val));
  });
}

wireSeg("countSeg", (v) => {
  state.count = v;
  $("customCount").classList.toggle("hidden", v !== 0);
});
wireSeg("workerSeg", (v) => (state.workers = v));
wireSeg("modeSeg", (v) => (state.mode = v));
wireSeg("timeoutSeg", (v) => (state.timeoutMs = v * 1000));

$("portInput").addEventListener("input", (e) => (state.port = Math.max(1, Number(e.target.value) || 443)));
$("sniInput").addEventListener("input", (e) => (state.sni = e.target.value.trim() || "speed.cloudflare.com"));
$("coloInput").addEventListener("input", (e) => (state.coloFilter = e.target.value.trim()));
$("speedToggle").addEventListener("change", (e) => (state.speedTest = e.target.checked));
$("wsToggle").addEventListener("change", (e) => (state.requireWS = e.target.checked));

// --- formatting helpers --------------------------------------------------------
const fmtMs = (v) => (Number.isFinite(v) && v > 0 ? v.toFixed(0) : "—");
const fmtSpeed = (b) => {
  if (!Number.isFinite(b) || b <= 0) return "—";
  return b >= 1048576 ? (b / 1048576).toFixed(1) + " MB/s" : (b / 1024).toFixed(0) + " KB/s";
};

function statePill(kind, text) {
  const el = $(kind === "scan" ? "scanStatePill" : "valStatePill");
  el.textContent = text;
  el.className = "pill " + (kind === "scan" ? "run" : "run");
}

function scanStateText() {
  return state.scanning ? "scanning…" : state.healthy.length ? "done" : "idle";
}

// --- stats ---------------------------------------------------------------------
function renderStats(s) {
  $("statsCard").style.display = "";
  $("stTested").textContent = s.tested;
  $("stHealthy").textContent = s.healthy;
  $("stFailed").textContent = s.failed;
  $("stInFlight").textContent = s.inFlight;
  const pct = s.total > 0 ? Math.min(100, Math.round((s.tested / s.total) * 100)) : 0;
  $("progressFill").style.width = pct + "%";
  $("progressLabel").textContent = `${s.tested} / ${s.total} (${pct}%)`;
}

// --- results tables -------------------------------------------------------------
function liveRow(r) {
  const tr = document.createElement("tr");
  tr.innerHTML = `
    <td>${r.ip}</td><td>${r.port}</td><td>${r.colo || "—"}</td>
    <td>${fmtMs(r.avgMs)}</td><td>${fmtMs(r.minMs)}</td>
    <td>${Number.isFinite(r.loss) ? r.loss.toFixed(0) + "%" : "—"}</td>
    <td>${fmtSpeed(r.throughput)}</td>
    <td class="${r.healthy ? "status-ok" : "status-bad"}">${r.healthy ? "healthy" : "failed"}</td>`;
  return tr;
}

function addLiveResult(r) {
  const body = $("resultsBody");
  body.prepend(liveRow(r));
  while (body.children.length > 300) body.removeChild(body.lastChild);
  if (r.healthy) addHealthy(r);
}

function addHealthy(r) {
  const idx = state.healthy.findIndex((h) => h.ip === r.ip && h.port === r.port);
  if (idx >= 0) state.healthy[idx] = r;
  else state.healthy.push(r);
  state.healthy.sort((a, b) => (a.avgMs || 1e9) - (b.avgMs || 1e9));
  renderHealthy();
}

function renderHealthy() {
  const body = $("healthyBody");
  body.innerHTML = "";
  state.healthy.forEach((h) => {
    const tr = document.createElement("tr");
    tr.innerHTML = `
      <td>${h.ip}</td><td>${h.port}</td><td>${h.colo || "—"}</td>
      <td>${fmtMs(h.avgMs)}</td><td>${fmtMs(h.minMs)}</td>
      <td>${Number.isFinite(h.loss) ? h.loss.toFixed(0) + "%" : "—"}</td>
      <td>${fmtSpeed(h.throughput)}</td>`;
    body.appendChild(tr);
  });
  $("healthyCountPill").textContent = state.healthy.length + " healthy";
  $("resultsCard").style.display = "";
  renderExportChips();
}

// --- validation ---------------------------------------------------------------
function addValidation(v) {
  const idx = state.validation.findIndex((x) => x.ip === v.ip && x.port === v.port);
  if (idx >= 0) state.validation[idx] = v;
  else state.validation.push(v);
  const body = $("validationBody");
  body.innerHTML = "";
  state.validation.forEach((x) => {
    const tr = document.createElement("tr");
    tr.innerHTML = `
      <td>${x.ip}</td><td>${x.port}</td>
      <td>${fmtMs(x.latencyMs)}</td>
      <td>${fmtSpeed(x.throughput)}</td>
      <td>${fmtSpeed(x.uploadThroughput)}</td>
      <td class="${x.success ? "status-ok" : "status-bad"}">${x.success ? "ok" : "failed"}</td>`;
    body.appendChild(tr);
  });
}

// --- export tab -----------------------------------------------------------------
function renderExportChips() {
  const wrap = $("endpointList");
  wrap.innerHTML = "";
  if (!state.healthy.length) {
    wrap.innerHTML = `<p class="empty">No working endpoints yet — run a scan first.</p>`;
  } else {
    state.healthy.forEach((h) => {
      const chip = document.createElement("span");
      chip.className = "endpoint-chip";
      chip.textContent = `${h.ip}:${h.port}`;
      wrap.appendChild(chip);
    });
  }
  $("exportCountPill").textContent = state.healthy.length;
}

function buildEndpoints() {
  return state.healthy.map((h) => `${h.ip}:${h.port}`);
}

function showBundle(bundle) {
  state.exported = bundle;
  $("subPreview").textContent = bundle.shareUrls.length ? bundle.shareUrls.join("\n") : "—";
  $("sbPreview").textContent = bundle.singBox ? bundle.singBox.slice(0, 2400) + (bundle.singBox.length > 2400 ? "\n… (truncated)" : "") : "—";
  $("exportCountPill").textContent = bundle.count;
}

// --- backend calls ----------------------------------------------------------------
const App = window.go?.main?.App;

async function refreshVersion() {
  try {
    $("versionChip").textContent = "v" + (await App.GetVersion()).replace(/^v/, "");
  } catch (e) {
    $("versionChip").textContent = "SenPai Scanner";
  }
}

function startScan() {
  if (state.scanning || !App) return;
  if (state.count === 0) {
    const custom = Math.max(1, Number($("customCount").value) || 0);
    if (!custom) return toast("Enter a custom IP count");
    state.count = custom;
  }
  const params = {
    count: state.count,
    workers: state.workers,
    timeoutMs: state.timeoutMs,
    tries: 4,
    port: state.port,
    mode: state.mode,
    sni: state.sni,
    speedTest: state.speedTest,
    requireWS: state.requireWS,
    coloFilter: state.coloFilter,
    outputFile: "",
  };
  state.liveResults = [];
  $("resultsBody").innerHTML = "";
  statePill("scan", "scanning…");
  $("startBtn").disabled = true;
  $("stopBtn").disabled = false;
  $("validateBtn").disabled = true;
  state.scanning = true;
  App.StartScan(params)
    .then(() => toast("Scan started — " + params.count.toLocaleString() + " IPs"))
    .catch((err) => {
      state.scanning = false;
      $("startBtn").disabled = false;
      $("stopBtn").disabled = true;
      statePill("scan", "idle");
      toast("Error: " + err);
    });
}

function stopScan() {
  if (!App || !state.scanning) return;
  App.StopScan();
  toast("Stopping…");
}

function validateTop() {
  if (!App) return;
  const url = $("configUrl").value.trim();
  if (!url) return toast("Paste a config URL first");
  const candidates = state.healthy.slice(0, Math.max(1, Number($("topNInput").value) || 10));
  if (!candidates.length) return toast("No healthy endpoints to validate");
  state.validation = [];
  $("validationBody").innerHTML = "";
  statePill("validate", "validating…");
  $("validateBtn").disabled = true;
  App.ValidateTopIPs(
    {
      configUrl: url,
      topN: candidates.length,
      timeoutMs: Math.max(1, Number($("valTimeoutInput").value) || 10) * 1000,
    },
    candidates
  )
    .then(() => toast(`Validating ${candidates.length} IPs through xray…`))
    .catch((err) => {
      statePill("validate", "idle");
      $("validateBtn").disabled = false;
      toast("Error: " + err);
    });
}

async function generateExport() {
  if (!App) return;
  const url = $("configUrl").value.trim();
  if (!url) return toast("Paste a config URL first");
  const eps = buildEndpoints();
  if (!eps.length) return toast("No working endpoints — run a scan first");
  try {
    const bundle = await App.GenerateConfigs(url, eps);
    showBundle(bundle);
    toast(`Generated ${bundle.count} configs`);
  } catch (err) {
    toast("Error: " + err);
  }
}

async function copyText(t) {
  if (!t) return toast("Nothing to copy yet");
  try {
    await App.CopyText(t);
    toast("Copied to clipboard");
  } catch (err) {
    toast("Copy failed: " + err);
  }
}

async function saveText(name, t) {
  if (!t) return toast("Nothing to save yet");
  try {
    const p = await App.SaveText(name, t);
    toast(p ? "Saved: " + p : "Save cancelled");
  } catch (err) {
    toast("Save failed: " + err);
  }
}

// --- event wiring ---------------------------------------------------------------
function wireEvents() {
  if (!window.runtime) return;
  window.runtime.EventsOn("scan:stats", (s) => renderStats(s));
  window.runtime.EventsOn("scan:result", (r) => addLiveResult(r));
  window.runtime.EventsOn("scan:done", (s) => {
    state.scanning = false;
    $("startBtn").disabled = false;
    $("stopBtn").disabled = true;
    $("validateBtn").disabled = !state.healthy.length;
    statePill("scan", "done");
    renderStats(s);
    toast(`Scan finished — ${state.healthy.length} healthy`);
  });
  window.runtime.EventsOn("scan:error", (e) => {
    state.scanning = false;
    $("startBtn").disabled = false;
    $("stopBtn").disabled = true;
    statePill("scan", "idle");
    toast("Scan error: " + e);
  });
  window.runtime.EventsOn("validate:result", (v) => addValidation(v));
  window.runtime.EventsOn("validate:done", (d) => {
    statePill("validate", "done");
    $("validateBtn").disabled = false;
    toast(`Validation done — ${d.count} checked`);
  });
  window.runtime.EventsOn("validate:error", (e) => {
    statePill("validate", "idle");
    $("validateBtn").disabled = false;
    toast("Validation error: " + e);
  });
}

// --- buttons -----------------------------------------------------------------------
$("startBtn").addEventListener("click", startScan);
$("stopBtn").addEventListener("click", stopScan);
$("validateBtn").addEventListener("click", validateTop);

$("copySubBtn").addEventListener("click", () => copyText(state.exported?.subscription));
$("saveSubBtn").addEventListener("click", () => saveText("senpai-subscription.txt", state.exported?.subscription));
$("copySbBtn").addEventListener("click", () => copyText(state.exported?.singBox));
$("saveSbBtn").addEventListener("click", () => saveText("senpai-singbox.json", state.exported?.singBox));
$("copyClashBtn").addEventListener("click", () => copyText(state.exported?.clash));
$("saveClashBtn").addEventListener("click", () => saveText("senpai-clash.yaml", state.exported?.clash));

// auto-generate export whenever the config URL changes while healthy IPs exist
let exportDebounce = null;
$("configUrl").addEventListener("input", () => {
  clearTimeout(exportDebounce);
  exportDebounce = setTimeout(() => {
    if (state.healthy.length && $("configUrl").value.trim()) generateExport();
  }, 600);
});

// --- boot -----------------------------------------------------------------------------
wireEvents();
refreshVersion();
