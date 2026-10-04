package com.matinsenpai.senpaiscanner.ui.main

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.matinsenpai.senpaiscanner.mobile.Callback
import com.matinsenpai.senpaiscanner.mobile.Mobile
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import kotlinx.serialization.Serializable
import kotlinx.serialization.decodeFromString
import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.Json

@Serializable
data class ScanConfig(
    val sourceType: String = "Random",
    val sourceFile: String = "",
    val countType: String = "5000",
    val customCount: String = "",
    val workerType: String = "50- default (restricted net)",
    val customWorkers: String = "",
    val timeoutType: String = "5s - default (restricted net)",
    val customTimeout: String = "",
    val portType: String = "Config",
    val selectedPorts: Set<Int> = setOf(443),
    val configUrl: String = "",
    val topNType: String = "50",
    val customTopN: String = "",
    val neighborScan: Boolean = false,
    val requireWebSocket: Boolean = true,
    val minSpeed: Double = 0.0,
    val speedUrl: String = "",
    val speedSize: Long = 524_288,
    val uploadTest: Boolean = false,
    // Gentle profile, pasted targets, skip-reachability and resume (same options as the desktop app)
    val gentle: Boolean = false,
    val targets: String = "",
    val phase2Only: Boolean = false,
    val resume: Boolean = false,
    // Anti-DPI recipe; empty fields are filled with the published values when the app starts
    val antiDpiEnabled: Boolean = true,
    val adFinalmask: String = "",
    val adFingerprint: String = "",
    val adAlpn: String = "",
    val adCiphers: String = "",
)

@Serializable
data class ResumeInfo(
    val saved: String = "",
    val tested: Int = 0,
    val total: Int = 0,
    val healthy: Int = 0,
    val phase: Int = 1,
)

@Serializable
data class TargetPreview(
    val count: Int = 0,
    val domains: Int = 0,
    val skipped: List<String> = emptyList(),
)

@Serializable
data class AntiDpiProfile(
    val finalmask: String = "",
    val fingerprint: String = "",
    val alpn: String = "",
    val cipherSuites: String = "",
)

data class IpResult(
    val ip: String,
    val port: Int,
    val latencyMs: Int,
    val loss: Double,
    val colo: String,
    val isHealthy: Boolean,
    val isPhase2: Boolean = false,
    val phase2Type: String = "",
    val phase2Speed: Double = 0.0,
    val phase2Status: Boolean = false,
    val phase2UploadSpeed: Double = 0.0,
)

@Serializable
data class ExportBundle(
    val subscription: String,
    val shareUrls: List<String>,
    val singBox: String,
    val clash: String,
    val count: Int,
)

@Serializable
data class MetaResponse(
    val as_organization: String = "",
    val ip: String = "",
    val colo: String = "",
)

enum class SessionPhase { IDLE, SCANNING, STOPPING, SPEED_TESTING }

data class ScanUiState(
    val phase: SessionPhase = SessionPhase.IDLE,
    val phase1Tested: Int = 0,
    val phase1Healthy: Int = 0,
    val phase1Failed: Int = 0,
    val phase1InFlight: Int = 0,
    val speedDone: Int = 0,
    val speedPassed: Int = 0,
    val speedFailed: Int = 0,
    val speedTotal: Int = 0,
    val results: List<IpResult> = emptyList(),
    val config: ScanConfig = ScanConfig(),
    val exportBundle: ExportBundle? = null,
    val error: String? = null,
    val isp: String = "",
    val publicIp: String = "",
    val networkColo: String = "",
    val isMetaLoading: Boolean = false,
    val resume: ResumeInfo? = null,
    val notice: String? = null,
    val targetsPreview: TargetPreview? = null,
    val antiDpiError: String? = null,
) {
    val isRunning: Boolean get() = phase != SessionPhase.IDLE
    val greenResults: List<IpResult>
        get() = results.filter { !it.isPhase2 && it.isHealthy }
    val speedResults: List<IpResult>
        get() = results.filter { it.isPhase2 }
}

class MainViewModel : ViewModel() {
    // encodeDefaults: fields left at their default (requireWebSocket, antiDpiEnabled, ...) must still reach the engine
    private val json = Json { ignoreUnknownKeys = true; encodeDefaults = true }
    private var previewJob: Job? = null
    private var antiDpiJob: Job? = null
    private val _uiState = MutableStateFlow(ScanUiState())
    val uiState: StateFlow<ScanUiState> = _uiState.asStateFlow()

    init {
        fetchUserMeta()
        refreshResume()
        loadAntiDpiDefaults(onlyIfBlank = true)
    }

    private val scanCallback = object : Callback {
        override fun onProgress(
            tested: Long,
            healthy: Long,
            failed: Long,
            inFlight: Long,
            isPhase2: Boolean,
        ) {
            _uiState.update { current ->
                if (isPhase2) {
                    val total = maxOf(current.speedTotal, tested.toInt() + inFlight.toInt())
                    current.copy(
                        speedDone = tested.toInt(),
                        speedPassed = healthy.toInt(),
                        speedFailed = failed.toInt(),
                        speedTotal = total,
                    )
                } else {
                    current.copy(
                        phase1Tested = tested.toInt(),
                        phase1Healthy = healthy.toInt(),
                        phase1Failed = failed.toInt(),
                        phase1InFlight = inFlight.toInt(),
                    )
                }
            }
        }

        override fun onResult(
            ip: String,
            port: Long,
            latencyMs: Long,
            loss: Double,
            colo: String,
            isHealthy: Boolean,
            isPhase2: Boolean,
            phase2Type: String,
            phase2Speed: Double,
            phase2Status: Boolean,
            phase2UploadSpeed: Double,
        ) {
            val incoming = IpResult(
                ip = ip,
                port = port.toInt(),
                latencyMs = latencyMs.toInt(),
                loss = loss,
                colo = colo,
                isHealthy = isHealthy,
                isPhase2 = isPhase2,
                phase2Type = phase2Type,
                phase2Speed = phase2Speed,
                phase2Status = phase2Status,
                phase2UploadSpeed = phase2UploadSpeed,
            )
            _uiState.update { current ->
                val withoutPrevious = current.results.filterNot {
                    it.ip == incoming.ip && it.port == incoming.port && it.isPhase2 == incoming.isPhase2
                }
                current.copy(results = listOf(incoming) + withoutPrevious, exportBundle = null)
            }
        }

        override fun onFinished() {
            _uiState.update { it.copy(phase = SessionPhase.IDLE, phase1InFlight = 0) }
            refreshResume()
        }

        override fun onInfo(msg: String) {
            _uiState.update { it.copy(notice = msg) }
        }

        override fun onError(err: String) {
            _uiState.update { it.copy(phase = SessionPhase.IDLE, phase1InFlight = 0, error = err) }
        }
    }

    fun updateConfig(config: ScanConfig) {
        _uiState.update { it.copy(config = config, exportBundle = null) }
    }

    fun fetchUserMeta() {
        viewModelScope.launch(Dispatchers.IO) {
            _uiState.update { it.copy(isMetaLoading = true) }
            runCatching {
                json.decodeFromString<MetaResponse>(Mobile.fetchMeta())
            }.onSuccess { meta ->
                _uiState.update {
                    it.copy(
                        isp = meta.as_organization,
                        publicIp = meta.ip,
                        networkColo = meta.colo,
                        isMetaLoading = false,
                    )
                }
            }.onFailure {
                _uiState.update { it.copy(isMetaLoading = false) }
            }
        }
    }

    fun startScan(resume: Boolean = false) {
        val current = _uiState.value
        if (current.isRunning) return
        _uiState.value = current.copy(
            phase = SessionPhase.SCANNING,
            phase1Tested = 0,
            phase1Healthy = 0,
            phase1Failed = 0,
            phase1InFlight = 0,
            speedDone = 0,
            speedPassed = 0,
            speedFailed = 0,
            speedTotal = 0,
            results = emptyList(),
            exportBundle = null,
            error = null,
            notice = null,
            resume = null,
        )
        Mobile.startScan(json.encodeToString(current.config.copy(resume = resume)), scanCallback)
    }

    fun resumeScan() = startScan(resume = true)

    /** Re-reads the saved-scan summary so the Scan tab can offer "Resume". */
    fun refreshResume() {
        viewModelScope.launch(Dispatchers.IO) {
            val payload = runCatching { Mobile.resumeInfo() }.getOrDefault("")
            val info = if (payload.isBlank()) null else runCatching { json.decodeFromString<ResumeInfo>(payload) }.getOrNull()
            _uiState.update { it.copy(resume = info) }
        }
    }

    fun discardResume() {
        viewModelScope.launch(Dispatchers.IO) {
            Mobile.discardResume()
            _uiState.update { it.copy(resume = null, notice = "The saved scan was discarded.") }
        }
    }

    /** Updates the pasted targets and, after a short pause, asks the engine how many addresses they contain. */
    fun onTargetsChanged(text: String) {
        _uiState.update { it.copy(config = it.config.copy(targets = text), exportBundle = null) }
        previewJob?.cancel()
        if (text.isBlank()) {
            _uiState.update { it.copy(targetsPreview = null) }
            return
        }
        previewJob = viewModelScope.launch(Dispatchers.IO) {
            delay(400)
            val preview = runCatching { json.decodeFromString<TargetPreview>(Mobile.previewTargets(text)) }.getOrNull()
            _uiState.update { it.copy(targetsPreview = preview) }
        }
    }

    /** Anti-DPI recipe edits go through here so the recipe is validated while the user types. */
    fun onAntiDpiChanged(config: ScanConfig) {
        updateConfig(config)
        antiDpiJob?.cancel()
        antiDpiJob = viewModelScope.launch(Dispatchers.IO) {
            delay(250)
            val problem = runCatching { Mobile.validateAntiDpi(json.encodeToString(config)) }.getOrDefault("")
            _uiState.update { it.copy(antiDpiError = problem.ifBlank { null }) }
        }
    }

    fun restoreAntiDpiDefaults() = loadAntiDpiDefaults(onlyIfBlank = false)

    private fun loadAntiDpiDefaults(onlyIfBlank: Boolean) {
        viewModelScope.launch(Dispatchers.IO) {
            val profile = runCatching { json.decodeFromString<AntiDpiProfile>(Mobile.antiDpiDefaults()) }.getOrNull() ?: return@launch
            _uiState.update { state ->
                val c = state.config
                val filled = if (onlyIfBlank) {
                    c.copy(
                        adFinalmask = c.adFinalmask.ifBlank { profile.finalmask },
                        adFingerprint = c.adFingerprint.ifBlank { profile.fingerprint },
                        adAlpn = c.adAlpn.ifBlank { profile.alpn },
                        adCiphers = c.adCiphers.ifBlank { profile.cipherSuites },
                    )
                } else {
                    c.copy(
                        antiDpiEnabled = true,
                        adFinalmask = profile.finalmask,
                        adFingerprint = profile.fingerprint,
                        adAlpn = profile.alpn,
                        adCiphers = profile.cipherSuites,
                    )
                }
                state.copy(
                    config = filled,
                    antiDpiError = null,
                    notice = if (onlyIfBlank) state.notice else "Suggested Anti-DPI values restored.",
                )
            }
        }
    }

    fun stop() {
        if (!_uiState.value.isRunning) return
        _uiState.update { it.copy(phase = SessionPhase.STOPPING) }
        Mobile.stopScan()
    }

    fun startSpeedTest() {
        val current = _uiState.value
        if (current.isRunning || current.greenResults.isEmpty()) return
        _uiState.value = current.copy(
            phase = SessionPhase.SPEED_TESTING,
            speedDone = 0,
            speedPassed = 0,
            speedFailed = 0,
            speedTotal = current.greenResults.size,
            results = current.results.filterNot { it.isPhase2 },
            exportBundle = null,
            error = null,
        )
        Mobile.startSpeedTest(json.encodeToString(current.config), scanCallback)
    }

    fun generateExports() {
        val current = _uiState.value
        val endpoints = current.speedResults
            .filter { it.phase2Status }
            .sortedBy { it.latencyMs }
            .map { "${it.ip}:${it.port}" }
            .distinct()
        if (current.config.configUrl.isBlank()) {
            _uiState.update { it.copy(error = "Add a VLESS, Trojan, or VMess URL in Scan first") }
            return
        }
        if (endpoints.isEmpty()) {
            _uiState.update { it.copy(error = "Run Speed Test and keep at least one passing endpoint first") }
            return
        }
        runCatching {
            val payload = Mobile.generateConfigs(current.config.configUrl, endpoints.joinToString("\n"))
            json.decodeFromString<ExportBundle>(payload)
        }.onSuccess { bundle ->
            _uiState.update { it.copy(exportBundle = bundle, error = null) }
        }.onFailure { failure ->
            _uiState.update { it.copy(error = failure.message ?: "Export failed") }
        }
    }

    fun dismissError() {
        _uiState.update { it.copy(error = null) }
    }

    fun dismissNotice() {
        _uiState.update { it.copy(notice = null) }
    }
}

fun healthyEndpoints(results: List<IpResult>, limit: Int = 0): List<IpResult> {
    val unique = results
        .asSequence()
        .filter { !it.isPhase2 && it.isHealthy }
        .sortedBy { it.latencyMs }
        .distinctBy { "${it.ip}:${it.port}" }
        .toList()
    return if (limit > 0) unique.take(limit) else unique
}
