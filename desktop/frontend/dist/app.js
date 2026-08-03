"use strict";

const App = window.go?.main?.App || null;
const Runtime = window.runtime || null;
const $ = (selector) => document.querySelector(selector);
const $$ = (selector) => [...document.querySelectorAll(selector)];

const state = {
  presets: null,
  activeTab: "scan",
  settings: {
    ipMode: 0,
    countIdx: 1, countCustom: "",
    workersIdx: 0, workersCustom: "",
    timeoutIdx: 2, timeoutCustom: "",
    ports: [],
    requireWS: true,
    neighborScan: false,
    topNIdx: 2, topNCustom: "",
    minSpeedIdx: 0, minSpeedCustom: "",
    speedSizeIdx: 1, speedSizeCustom: "",
    uploadTest: false,
  },
  scan: { running: false, phase: 0, cancelled: false, manualSpeed: false, livePath: "", status: "idle" },
  stats: { tested: 0, healthy: 0, failed: 0, total: 0 },
  phase2: { done: 0, total: 0 },
  results: [],
  validation: new Map(),
  workingEndpoints: [],
  resultFilter: "healthy",
  resultSearch: "",
  sortKey: "avg",
  exportBundle: null,
};

const D = { STATS: 1, RESULTS: 2, SPEED: 4, ACTIONS: 8, EXPORT: 16 };
let dirty = 0;
let renderQueued = false;
function mark(...bits) {
  bits.forEach((bit) => { dirty |= bit; });
  if (!renderQueued) {
    renderQueued = true;
    requestAnimationFrame(renderDirty);
  }
}

const els = {
  sessionRail: $("#sessionRail"), phaseTitle: $("#phaseTitle"), phaseHint: $("#phaseHint"),
  progressText: $("#progressText"), progressPercent: $("#progressPercent"), progressTrack: $("#progressTrack"), barFill: $("#barFill"),
  statTested: $("#statTested"), statHealthy: $("#statHealthy"), statFailed: $("#statFailed"), statPhase2: $("#statPhase2"),
  btnStart: $("#btnStart"), btnStartBottom: $("#btnStartBottom"), btnStop: $("#btnStop"), btnRetry: $("#btnRetry"),
  btnCopyGreen: $("#btnCopyGreen"), btnCopyTop20: $("#btnCopyTop20"), btnSpeedGreen: $("#btnSpeedGreen"),
  btnCopyGreenExport: $("#btnCopyGreenExport"), btnCopyTop20Export: $("#btnCopyTop20Export"), btnSaveGreen: $("#btnSaveGreen"),
  btnBuildExport: $("#btnBuildExport"), btnExportDisk: $("#btnExportDisk"),
  resultsBody: $("#resultsBody"), validationBody: $("#validationBody"), resultsEmpty: $("#resultsEmpty"), speedEmpty: $("#speedEmpty"),
  resultsTabCount: $("#resultsTabCount"), exportTabCount: $("#exportTabCount"), greenFilterCount: $("#greenFilterCount"), allFilterCount: $("#allFilterCount"),
  speedResultBadge: $("#speedResultBadge"), rawExportCount: $("#rawExportCount"), resultLimitNote: $("#resultLimitNote"),
  livePathPill: $("#livePathPill"), livePathText: $("#livePathText"),
  ispChip: $("#ispChip"), ispText: $("#ispText"), networkDetail: $("#networkDetail"), versionText: $("#versionText"),
  configUrl: $("#configUrl"), speedUrl: $("#speedUrl"), portChips: $("#portChips"),
  countCustom: $("#countCustom"), workersCustom: $("#workersCustom"), timeoutCustom: $("#timeoutCustom"),
  topNCustom: $("#topNCustom"), minSpeedCustom: $("#minSpeedCustom"), speedSizeCustom: $("#speedSizeCustom"),
  toggleRequireWS: $("#toggleRequireWS"), toggleNeighbors: $("#toggleNeighbors"), toggleUpload: $("#toggleUpload"),
  exportSub: $("#exportSub"), exportSingbox: $("#exportSingbox"), exportClash: $("#exportClash"), exportNote: $("#exportNote"),
  exportCallout: $("#exportCallout"), subCount: $("#subCount"), toast: $("#toast"),
};

let toastTimer;
function toast(message) {
  els.toast.textContent = String(message);
  els.toast.classList.add("show");
  clearTimeout(toastTimer);
  toastTimer = setTimeout(() => els.toast.classList.remove("show"), 2800);
}

async function invoke(fn, ...args) {
  if (!fn) {
    toast("Backend is not ready.");
    return { ok: false, value: null };
  }
  try {
    return { ok: true, value: await fn(...args) };
  } catch (error) {
    toast(error?.message || error || "The action failed.");
    return { ok: false, value: null };
  }
}

function parseDurationMs(value) {
  const match = String(value || "").trim().match(/^(\d+(?:\.\d+)?)\s*(ms|s|m)?$/i);
  if (!match) return 0;
  const amount = Number(match[1]);
  const unit = (match[2] || "s").toLowerCase();
  return Math.round(amount * (unit === "ms" ? 1 : unit === "m" ? 60000 : 1000));
}

function fmtSpeed(bytesPerSecond) {
  if (!(bytesPerSecond > 0)) return "—";
  const mbps = bytesPerSecond * 8 / 1e6;
  return mbps >= 1 ? `${mbps.toFixed(1)} Mbps` : `${Math.round(mbps * 1000)} Kbps`;
}
function fmtMs(value) { return value > 0 ? `${Math.round(value)} ms` : "—"; }
function unique(values) { return [...new Set(values)]; }
function endpointOf(result) { return `${result.ip}:${result.port}`; }

const SEGS = {
  count: { idx: "countIdx", custom: "countCustom", labels: "countLabels", values: "countValues" },
  workers: { idx: "workersIdx", custom: "workersCustom", labels: "workerLabels", values: "workerValues" },
  timeout: { idx: "timeoutIdx", custom: "timeoutCustom", labels: "timeoutLabels", values: "timeoutValues" },
  topN: { idx: "topNIdx", custom: "topNCustom", labels: "topNLabels", values: "topNValues" },
  minSpeed: { idx: "minSpeedIdx", custom: "minSpeedCustom", labels: "minSpeedLabels", values: "minSpeedValues" },
  speedSize: { idx: "speedSizeIdx", custom: "speedSizeCustom", labels: "speedSizeLabels", values: "speedSizeValues" },
};

function buildSegments() {
  const ipMode = $("[data-seg='ipMode']");
  ipMode.querySelectorAll("button").forEach((button) => {
    button.onclick = () => {
      state.settings.ipMode = Number(button.dataset.val) || 0;
      ipMode.querySelectorAll("button").forEach((item) => item.classList.toggle("on", item === button));
    };
  });
  ipMode.querySelectorAll("button").forEach((button) => button.classList.toggle("on", Number(button.dataset.val) === state.settings.ipMode));

  Object.entries(SEGS).forEach(([key, def]) => {
    const root = $(`[data-seg='${key}']`);
    root.replaceChildren();
    state.presets[def.labels].forEach((label, index) => {
      const button = document.createElement("button");
      button.type = "button";
      button.textContent = label;
      button.classList.toggle("on", state.settings[def.idx] === index);
      button.onclick = () => {
        state.settings[def.idx] = index;
        root.querySelectorAll("button").forEach((item, i) => item.classList.toggle("on", i === index));
        syncCustomField(key);
      };
      root.appendChild(button);
    });
    syncCustomField(key);
  });
}

function syncCustomField(key) {
  const def = SEGS[key];
  const input = els[def.custom];
  const isCustom = state.settings[def.idx] === state.presets[def.labels].length - 1;
  input.classList.toggle("hidden", !isCustom);
}

function buildPorts() {
  els.portChips.replaceChildren();
  state.presets.ports.forEach((port) => {
    const button = document.createElement("button");
    button.type = "button";
    button.className = "chip";
    button.dataset.port = String(port);
    button.textContent = port === 0 ? "Config" : String(port);
    button.onclick = () => {
      if (port === 0) state.settings.ports = [];
      else if (state.settings.ports.includes(port)) state.settings.ports = state.settings.ports.filter((item) => item !== port);
      else state.settings.ports.push(port);
      renderPorts();
    };
    els.portChips.appendChild(button);
  });
  renderPorts();
}

function renderPorts() {
  els.portChips.querySelectorAll("button").forEach((button) => {
    const port = Number(button.dataset.port);
    button.classList.toggle("on", port === 0 ? state.settings.ports.length === 0 : state.settings.ports.includes(port));
  });
}

function setToggle(element, enabled) {
  element.classList.toggle("on", !!enabled);
  element.setAttribute("aria-checked", enabled ? "true" : "false");
}

function resolvedPreset(def, customValue, parser, fallback) {
  const index = state.settings[def.idx];
  const values = state.presets[def.values];
  const labels = state.presets[def.labels];
  const raw = index === labels.length - 1 ? customValue : values[index];
  const value = parser(raw);
  return Number.isFinite(value) ? value : fallback;
}

function readSettings() {
  const count = Math.max(1, resolvedPreset(SEGS.count, state.settings.countCustom, (v) => parseInt(v, 10), 1000));
  const workers = Math.max(1, resolvedPreset(SEGS.workers, state.settings.workersCustom, (v) => parseInt(v, 10), 50));
  const timeoutMs = Math.max(1, resolvedPreset(SEGS.timeout, state.settings.timeoutCustom, parseDurationMs, 5000));
  const topN = Math.max(0, resolvedPreset(SEGS.topN, state.settings.topNCustom, (v) => parseInt(v, 10), 50));
  const minSpeed = Math.max(0, resolvedPreset(SEGS.minSpeed, state.settings.minSpeedCustom, (v) => parseFloat(v), 0));
  let speedSize = resolvedPreset(SEGS.speedSize, state.settings.speedSizeCustom, (v) => parseInt(v, 10), 512 * 1024);
  if (state.settings.speedSizeIdx === state.presets.speedSizeLabels.length - 1) speedSize = Math.max(1, parseFloat(state.settings.speedSizeCustom) || .5) * 1024 * 1024;
  return {
    ipMode: state.settings.ipMode,
    count, workers, timeoutMs,
    ports: state.settings.ports.length ? [...state.settings.ports] : [0],
    configUrl: els.configUrl.value.trim(),
    requireWS: state.settings.requireWS,
    neighborScan: state.settings.neighborScan,
    topN, minSpeed,
    speedUrl: els.speedUrl.value.trim(),
    speedSize: Math.round(speedSize),
    uploadTest: state.settings.uploadTest,
    countIdx: state.settings.countIdx, countCustom: state.settings.countCustom,
    workersIdx: state.settings.workersIdx, workersCustom: state.settings.workersCustom,
    timeoutIdx: state.settings.timeoutIdx, timeoutCustom: state.settings.timeoutCustom,
    topNIdx: state.settings.topNIdx, topNCustom: state.settings.topNCustom,
    minSpeedIdx: state.settings.minSpeedIdx, minSpeedCustom: state.settings.minSpeedCustom,
    speedSizeIdx: state.settings.speedSizeIdx, speedSizeCustom: state.settings.speedSizeCustom,
  };
}

function applyParams(params) {
  const s = state.settings;
  ["countIdx", "workersIdx", "timeoutIdx", "topNIdx", "minSpeedIdx", "speedSizeIdx"].forEach((key) => { s[key] = Number(params[key]) || 0; });
  ["countCustom", "workersCustom", "timeoutCustom", "topNCustom", "minSpeedCustom", "speedSizeCustom"].forEach((key) => { s[key] = params[key] || ""; els[key].value = s[key]; });
  s.ipMode = Number(params.ipMode) || 0;
  s.ports = (params.ports || []).filter((port) => port > 0);
  s.requireWS = params.requireWS !== false;
  s.neighborScan = !!params.neighborScan;
  s.uploadTest = !!params.uploadTest;
  els.configUrl.value = params.configUrl || "";
  els.speedUrl.value = params.speedUrl || "";
  buildSegments();
  renderPorts();
  setToggle(els.toggleRequireWS, s.requireWS);
  setToggle(els.toggleNeighbors, s.neighborScan);
  setToggle(els.toggleUpload, s.uploadTest);
}

function sortRank(result) {
  if (result.healthy) return 0;
  if (result.avgMs > 0 || result.loss < 100) return 1;
  return 2;
}

function sortResults(list) {
  return [...list].sort((a, b) => {
    const rank = sortRank(a) - sortRank(b);
    if (rank) return rank;
    if (state.sortKey === "loss") return a.loss - b.loss || a.avgMs - b.avgMs;
    if (state.sortKey === "jitter") return a.jitterMs - b.jitterMs || a.avgMs - b.avgMs;
    if (state.sortKey === "colo") return (a.colo || "").localeCompare(b.colo || "") || a.avgMs - b.avgMs;
    if (state.sortKey === "speed") return (b.throughput || 0) - (a.throughput || 0) || a.avgMs - b.avgMs;
    return a.avgMs - b.avgMs;
  });
}

function healthyResults() { return sortResults(state.results.filter((result) => result.healthy)); }
function greenEndpoints() { return unique(healthyResults().map(endpointOf)); }
function top20IPs() { return unique(healthyResults().map((result) => result.ip)).slice(0, 20); }

function appendCell(row, text, className = "") {
  const cell = document.createElement("td");
  cell.textContent = text;
  if (className) cell.className = className;
  row.appendChild(cell);
}

function renderResults() {
  const query = state.resultSearch.trim().toLowerCase();
  let rows = state.resultFilter === "healthy" ? healthyResults() : sortResults(state.results);
  if (query) rows = rows.filter((result) => `${result.ip}:${result.port} ${result.colo || ""}`.toLowerCase().includes(query));
  const cap = 500;
  const visible = rows.slice(0, cap);
  const fragment = document.createDocumentFragment();
  visible.forEach((result) => {
    const row = document.createElement("tr");
    appendCell(row, endpointOf(result), "endpoint");
    appendCell(row, result.colo || "—");
    appendCell(row, fmtMs(result.avgMs), "num");
    appendCell(row, `${Math.round(result.loss)}%`, `num ${result.loss === 0 ? "good" : result.loss >= 100 ? "bad" : ""}`);
    appendCell(row, fmtMs(result.jitterMs), "num");
    appendCell(row, fmtSpeed(result.throughput), "num");
    appendCell(row, result.healthy ? "green" : "failed", `status-mark ${result.healthy ? "good" : "bad"}`);
    fragment.appendChild(row);
  });
  els.resultsBody.replaceChildren(fragment);
  els.resultsEmpty.classList.toggle("hidden", visible.length > 0);
  els.resultLimitNote.textContent = rows.length > cap ? `Showing the best ${cap.toLocaleString()} of ${rows.length.toLocaleString()} matching rows. Copy actions use the complete set.` : "";
  const greenCount = greenEndpoints().length;
  els.resultsTabCount.textContent = greenCount;
  els.greenFilterCount.textContent = greenCount;
  els.allFilterCount.textContent = state.results.length;
  els.rawExportCount.textContent = `${greenCount} green`;
}

function renderSpeedResults() {
  const outcomes = [...state.validation.values()].sort((a, b) => Number(b.success) - Number(a.success) || (b.throughput || 0) - (a.throughput || 0));
  const fragment = document.createDocumentFragment();
  outcomes.forEach((outcome) => {
    const row = document.createElement("tr");
    appendCell(row, `${outcome.ip}:${outcome.port}`, "endpoint");
    appendCell(row, outcome.transport || "—");
    appendCell(row, fmtSpeed(outcome.throughput), "num");
    appendCell(row, fmtSpeed(outcome.uploadThroughput), "num");
    appendCell(row, fmtMs(outcome.latencyMs), "num");
    appendCell(row, outcome.success ? "working" : "failed", `status-mark ${outcome.success ? "good" : "bad"}`);
    appendCell(row, outcome.error || "—");
    fragment.appendChild(row);
  });
  els.validationBody.replaceChildren(fragment);
  els.speedEmpty.classList.toggle("hidden", outcomes.length > 0);
  els.speedResultBadge.textContent = `${outcomes.length} tested`;
  els.exportTabCount.textContent = state.workingEndpoints.length;
}

function renderStats() {
  els.statTested.textContent = state.stats.tested.toLocaleString();
  els.statHealthy.textContent = state.stats.healthy.toLocaleString();
  els.statFailed.textContent = state.stats.failed.toLocaleString();
  els.statPhase2.textContent = `${state.phase2.done} / ${state.phase2.total || "—"}`;
  const phase2Active = state.scan.phase === 2 && state.phase2.total > 0;
  const current = phase2Active ? state.phase2.done : state.stats.tested;
  const total = phase2Active ? state.phase2.total : state.stats.total;
  const ratio = total > 0 ? Math.min(1, current / total) : 0;
  const percent = Math.round(ratio * 100);
  els.barFill.style.transform = `scaleX(${ratio})`;
  els.progressPercent.textContent = `${percent}%`;
  els.progressTrack.setAttribute("aria-valuenow", String(percent));
  if (state.scan.status === "idle") els.progressText.textContent = "No active run";
  else if (phase2Active) els.progressText.textContent = `${current.toLocaleString()} of ${total.toLocaleString()} speed tests`;
  else if (total > 0) els.progressText.textContent = `${current.toLocaleString()} of ${total.toLocaleString()} probes${current > total ? " + neighbors" : ""}`;
  else els.progressText.textContent = "Preparing targets";
}

function canGenerateConfigs() { return els.configUrl.value.trim() !== "" && state.workingEndpoints.length > 0; }
function syncActions() {
  const greenCount = greenEndpoints().length;
  const hasGreen = greenCount > 0;
  const running = state.scan.running;
  els.btnStart.disabled = running;
  els.btnStartBottom.disabled = running;
  els.btnRetry.disabled = running;
  els.btnStop.disabled = !running;
  [els.btnCopyGreen, els.btnCopyTop20, els.btnCopyGreenExport, els.btnCopyTop20Export, els.btnSaveGreen].forEach((button) => { button.disabled = !hasGreen; });
  els.btnSpeedGreen.disabled = running || !hasGreen;
  els.btnSpeedGreen.textContent = hasGreen ? `Speed test ${greenCount} green result${greenCount === 1 ? "" : "s"}` : "Speed test green results";
  const generate = canGenerateConfigs();
  els.btnBuildExport.disabled = !generate || running;
  els.btnExportDisk.disabled = !generate || running;
  if (generate) {
    els.exportCallout.classList.add("ready");
    els.exportCallout.querySelector("strong").textContent = `${state.workingEndpoints.length} working endpoint${state.workingEndpoints.length === 1 ? "" : "s"} ready.`;
    els.exportCallout.querySelector("span").textContent = "Generate the configuration pack or save all formats beside the app.";
  } else {
    els.exportCallout.classList.remove("ready");
    els.exportCallout.querySelector("strong").textContent = "Run a tunnel speed test first.";
    els.exportCallout.querySelector("span").textContent = "A share URL and at least one working endpoint are required to generate client configs.";
  }
}

function renderExport() {
  const bundle = state.exportBundle;
  els.exportSub.value = bundle?.subscription || "";
  els.exportSingbox.value = bundle?.singBox || "";
  els.exportClash.value = bundle?.clash || "";
  els.subCount.textContent = bundle ? `${bundle.count} endpoints` : "empty";
  els.exportNote.textContent = bundle ? `${bundle.count} endpoint${bundle.count === 1 ? "" : "s"} generated from the current share URL.` : "";
}

function renderDirty() {
  renderQueued = false;
  const bits = dirty;
  dirty = 0;
  if (bits & D.STATS) renderStats();
  if (bits & D.RESULTS) renderResults();
  if (bits & D.SPEED) renderSpeedResults();
  if (bits & D.ACTIONS) syncActions();
  if (bits & D.EXPORT) renderExport();
}

function switchTab(tab) {
  state.activeTab = tab;
  $$(".tab").forEach((button) => {
    const active = button.dataset.tab === tab;
    button.classList.toggle("active", active);
    button.setAttribute("aria-selected", active ? "true" : "false");
    button.tabIndex = active ? 0 : -1;
  });
  $$("[data-panel]").forEach((panel) => {
    const active = panel.dataset.panel === tab;
    panel.classList.toggle("active", active);
    panel.hidden = !active;
  });
  $(".workspace").scrollTop = 0;
}

function setSession(status, title, hint) {
  state.scan.status = status;
  els.sessionRail.dataset.state = status;
  els.phaseTitle.textContent = title;
  els.phaseHint.textContent = hint;
}

function resetRun(params) {
  state.results = [];
  state.validation.clear();
  state.workingEndpoints = [];
  state.exportBundle = null;
  state.stats = { tested: 0, healthy: 0, failed: 0, total: params.count * Math.max(1, params.ports.length) };
  state.phase2 = { done: 0, total: 0 };
  state.scan = { running: true, phase: 1, cancelled: false, manualSpeed: false, livePath: "", status: "running" };
  setSession("running", "Phase 1 · reachability", "Green results are available to copy immediately.");
  mark(D.STATS, D.RESULTS, D.SPEED, D.ACTIONS, D.EXPORT);
}

async function startScan(params = readSettings()) {
  const response = await invoke(App?.StartScan, params);
  if (!response.ok) return;
  resetRun(params);
  switchTab("results");
}

async function retryLast() {
  const response = await invoke(App?.RetryLastScan);
  if (!response.ok || !response.value) return;
  applyParams(response.value);
  toast("Loaded the last scan settings.");
  await startScan(response.value);
}

async function stopScan() {
  if (!state.scan.running) return;
  els.btnStop.disabled = true;
  setSession("running", "Stopping safely…", "Finishing in-flight probes and keeping every green result.");
  await invoke(App?.StopScan);
}

async function speedTestGreen() {
  if (state.scan.running || greenEndpoints().length === 0) return;
  state.validation.clear();
  state.workingEndpoints = [];
  state.exportBundle = null;
  state.phase2 = { done: 0, total: greenEndpoints().length };
  state.scan.running = true;
  state.scan.phase = 2;
  state.scan.manualSpeed = true;
  setSession("running", "Speed testing current green results", els.configUrl.value.trim() ? "Traffic is routed through the current share URL." : "Direct Cloudflare download samples are running.");
  mark(D.STATS, D.SPEED, D.ACTIONS, D.EXPORT);
  const response = await invoke(App?.StartSpeedTest, readSettings());
  if (!response.ok) {
    state.scan.running = false;
    setSession("stopped", "Scan stopped", "Green results are preserved and ready to copy.");
    mark(D.ACTIONS);
    return;
  }
  switchTab("results");
}

async function onScanDone(payload = {}) {
  state.scan.running = false;
  state.scan.cancelled = !!payload.cancelled;
  state.workingEndpoints = payload.workingEndpoints || [];
  if (payload.manualSpeed) {
    setSession(payload.cancelled ? "stopped" : "done", payload.cancelled ? "Speed test stopped" : "Speed test complete", `${state.workingEndpoints.length} endpoint${state.workingEndpoints.length === 1 ? "" : "s"} passed.`);
    if (!payload.cancelled && canGenerateConfigs()) await buildExport(false);
    toast(payload.cancelled ? "Speed test stopped; completed results were kept." : `Speed test complete · ${state.workingEndpoints.length} working.`);
  } else if (payload.cancelled) {
    setSession("stopped", "Scan stopped", `${greenEndpoints().length} green endpoint${greenEndpoints().length === 1 ? " is" : "s are"} ready to copy or speed test.`);
    toast("Scan stopped. Current green results were preserved.");
  } else {
    setSession("done", "Scan complete", els.configUrl.value.trim() ? `${state.workingEndpoints.length} endpoint${state.workingEndpoints.length === 1 ? "" : "s"} passed tunnel validation.` : `${greenEndpoints().length} green endpoint${greenEndpoints().length === 1 ? "" : "s"} found.`);
    if (canGenerateConfigs()) await buildExport(false);
    toast("Scan complete.");
  }
  mark(D.STATS, D.RESULTS, D.SPEED, D.ACTIONS, D.EXPORT);
}

async function copyText(text, successMessage) {
  if (!text) { toast("Nothing is available yet."); return; }
  const response = await invoke(App?.CopyText, text);
  if (response.ok) toast(successMessage);
}

async function buildExport(showToast = true) {
  if (!canGenerateConfigs()) { toast("A share URL and working speed-test results are required."); return; }
  const response = await invoke(App?.GenerateConfigs, els.configUrl.value.trim(), state.workingEndpoints);
  if (!response.ok || !response.value) return;
  state.exportBundle = response.value;
  mark(D.EXPORT, D.ACTIONS);
  if (showToast) toast(`Generated ${response.value.count} client configuration${response.value.count === 1 ? "" : "s"}.`);
}

function wireEvents() {
  if (!Runtime?.EventsOn) return;
  Runtime.EventsOn("app:meta", (meta) => {
    els.ispChip.classList.add("ready");
    els.ispText.textContent = meta.asOrganization || "Unknown ISP";
    const details = [];
    if (meta.asn) details.push(`AS${meta.asn}`);
    if (meta.country) details.push(meta.country);
    if (meta.colo) details.push(`colo ${meta.colo}`);
    els.networkDetail.textContent = details.join(" · ") || meta.ip || "public network";
    els.ispChip.title = [meta.ip, meta.source].filter(Boolean).join(" · ");
  });
  Runtime.EventsOn("scan:phase", (phase) => {
    state.scan.phase = phase.phase;
    state.scan.manualSpeed = !!phase.manual;
    if (phase.livePath) {
      state.scan.livePath = phase.livePath;
      els.livePathPill.querySelector("span").textContent = phase.livePath;
      els.livePathText.textContent = phase.livePath;
    }
    if (phase.phase === 2) {
      const direct = phase.mode === "direct";
      setSession("running", phase.manual ? "Speed testing current green results" : "Phase 2 · tunnel validation", direct ? "Direct Cloudflare download samples are running." : "Testing the selected endpoints through xray.");
      state.phase2.done = 0;
      if (phase.manual) state.phase2.total = greenEndpoints().length;
      mark(D.STATS, D.ACTIONS);
    }
  });
  Runtime.EventsOn("scan:stats", (stats) => {
    state.stats.tested = stats.tested || 0;
    state.stats.healthy = stats.healthy || 0;
    state.stats.failed = stats.failed || 0;
    if (stats.total) state.stats.total = stats.total;
    mark(D.STATS);
  });
  Runtime.EventsOn("scan:results", (batch) => {
    if (!Array.isArray(batch)) return;
    state.results.push(...batch);
    mark(D.RESULTS, D.ACTIONS);
  });
  Runtime.EventsOn("validate:result", (outcome) => {
    state.validation.set(`${outcome.ip}:${outcome.port}`, outcome);
    state.phase2.done = Math.max(state.phase2.done, outcome.done || 0);
    state.phase2.total = outcome.total || state.phase2.total;
    mark(D.SPEED, D.STATS);
  });
  Runtime.EventsOn("scan:done", onScanDone);
  Runtime.EventsOn("scan:error", (message) => toast(message));
}

function wireControls() {
  $$(".tab").forEach((tab) => {
    tab.onclick = () => switchTab(tab.dataset.tab);
    tab.onkeydown = (event) => {
      if (!["ArrowLeft", "ArrowRight"].includes(event.key)) return;
      const tabs = $$(".tab");
      const next = (tabs.indexOf(tab) + (event.key === "ArrowRight" ? 1 : -1) + tabs.length) % tabs.length;
      tabs[next].focus(); tabs[next].click();
    };
  });
  els.btnStart.onclick = () => startScan();
  els.btnStartBottom.onclick = () => startScan();
  els.btnRetry.onclick = retryLast;
  els.btnStop.onclick = stopScan;
  els.btnSpeedGreen.onclick = speedTestGreen;

  [[els.toggleRequireWS, "requireWS"], [els.toggleNeighbors, "neighborScan"], [els.toggleUpload, "uploadTest"]].forEach(([element, key]) => {
    element.onclick = () => { state.settings[key] = !state.settings[key]; setToggle(element, state.settings[key]); };
  });
  Object.values(SEGS).forEach((def) => {
    els[def.custom].oninput = (event) => { state.settings[def.custom] = event.target.value; };
  });

  $$(".filter").forEach((button) => { button.onclick = () => { state.resultFilter = button.dataset.filter; $$(".filter").forEach((item) => item.classList.toggle("active", item === button)); mark(D.RESULTS); }; });
  $$(".sort").forEach((button) => { button.onclick = () => { state.sortKey = button.dataset.sort; $$(".sort").forEach((item) => item.classList.toggle("active", item === button)); mark(D.RESULTS); }; });
  $("#resultSearch").oninput = (event) => { state.resultSearch = event.target.value; mark(D.RESULTS); };

  const copyGreen = () => copyText(greenEndpoints().join("\n"), `Copied ${greenEndpoints().length} green endpoints.`);
  const copyTop = () => copyText(top20IPs().join("\n"), `Copied ${top20IPs().length} top IPs.`);
  els.btnCopyGreen.onclick = copyGreen;
  els.btnCopyGreenExport.onclick = copyGreen;
  els.btnCopyTop20.onclick = copyTop;
  els.btnCopyTop20Export.onclick = copyTop;
  els.btnSaveGreen.onclick = async () => {
    const text = greenEndpoints().join("\n");
    if (!text) return;
    const response = await invoke(App?.SaveText, "senpaiscanner-green-endpoints.txt", text);
    if (response.ok && response.value) toast(`Saved to ${response.value}`);
  };
  const copyLivePath = () => copyText(state.scan.livePath, "Copied the live results path.");
  els.livePathPill.onclick = copyLivePath;
  els.livePathPill.onkeydown = (event) => { if (event.key === "Enter" || event.key === " ") copyLivePath(); };
  els.btnBuildExport.onclick = () => buildExport(true);
  els.btnExportDisk.onclick = async () => {
    const response = await invoke(App?.ExportAllToDisk, els.configUrl.value.trim(), state.workingEndpoints);
    if (response.ok && response.value) toast(`Saved all formats to ${response.value}`);
  };

  const formats = [
    ["exportSub", "senpaiscanner-sub.txt", "btnCopySub", "btnSaveSub"],
    ["exportSingbox", "senpaiscanner-singbox.json", "btnCopySingbox", "btnSaveSingbox"],
    ["exportClash", "senpaiscanner-clash.yaml", "btnCopyClash", "btnSaveClash"],
  ];
  formats.forEach(([field, filename, copyID, saveID]) => {
    $(`#${copyID}`).onclick = () => copyText(els[field].value, "Copied to the clipboard.");
    $(`#${saveID}`).onclick = async () => {
      if (!els[field].value) { toast("Generate the export pack first."); return; }
      const response = await invoke(App?.SaveText, filename, els[field].value);
      if (response.ok && response.value) toast(`Saved to ${response.value}`);
    };
  });
  els.configUrl.oninput = () => mark(D.ACTIONS);
}

async function init() {
  const fallbackPresets = {
    countLabels: ["1,000", "5,000", "20,000", "Custom"], countValues: ["1000", "5000", "20000", "0"],
    workerLabels: ["50", "100", "200", "Custom"], workerValues: ["50", "100", "200", ""],
    timeoutLabels: ["2s", "3s", "5s", "Custom"], timeoutValues: ["2s", "3s", "5s", ""],
    topNLabels: ["10", "25", "50", "100", "All", "Custom"], topNValues: ["10", "25", "50", "100", "0", "0"],
    minSpeedLabels: ["None", "1 Mbps", "2 Mbps", "5 Mbps", "Custom"], minSpeedValues: ["0", "1", "2", "5", "-1"],
    speedSizeLabels: ["128 KB", "512 KB", "1 MB", "5 MB", "Custom"], speedSizeValues: ["131072", "524288", "1048576", "5242880", "0"],
    ports: [0, 443, 8443, 2053, 2083, 2087, 2096],
  };
  if (App) {
    const [version, presets] = await Promise.all([invoke(App.GetVersion), invoke(App.Presets)]);
    els.versionText.textContent = version.value || "v0.8.0";
    state.presets = presets.value || fallbackPresets;
  } else {
    els.versionText.textContent = "v0.8.0 preview";
    state.presets = fallbackPresets;
  }
  buildSegments();
  buildPorts();
  setToggle(els.toggleRequireWS, state.settings.requireWS);
  setToggle(els.toggleNeighbors, state.settings.neighborScan);
  setToggle(els.toggleUpload, state.settings.uploadTest);
  wireControls();
  if (App) wireEvents();
  else toast("Preview mode · launch through Wails to scan.");
  const previewTab = location.hash.slice(1);
  if (["scan", "results", "export"].includes(previewTab)) switchTab(previewTab);
  mark(D.STATS, D.RESULTS, D.SPEED, D.ACTIONS, D.EXPORT);
}

document.addEventListener("DOMContentLoaded", init);
