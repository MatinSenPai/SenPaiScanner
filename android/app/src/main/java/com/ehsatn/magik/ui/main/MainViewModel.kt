package com.ehsatn.magik.ui.main

import androidx.lifecycle.ViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.serialization.Serializable
import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.Json
import com.ehsatn.magik.mobile.Callback
import com.ehsatn.magik.mobile.Mobile

@Serializable
data class ScanConfig(
    // Source
    val sourceType: String = "Random",
    val sourceFile: String = "",

    // Count
    val countType: String = "5000",
    val customCount: String = "",
    
    // Workers
    val workerType: String = "50- default (restricted net)",
    val customWorkers: String = "",

    // Timeout
    val timeoutType: String = "5s - default (restricted net)",
    val customTimeout: String = "",

    // Ports
    val portType: String = "Config", // "Config" or "CustomPorts"
    val selectedPorts: Set<Int> = emptySet(),
    val configUrl: String = "",

    // Top N
    val topNType: String = "50",
    val customTopN: String = ""
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
    val phase2Status: Boolean = false
)

data class ScanUiState(
    val isRunning: Boolean = false,
    val tested: Int = 0,
    val healthy: Int = 0,
    val failed: Int = 0,
    val inFlight: Int = 0,
    val isPhase2: Boolean = false,
    val totalPhase2: Int = 0,
    val results: List<IpResult> = emptyList(),
    val error: String? = null,
    val config: ScanConfig = ScanConfig()
)

class MainViewModel : ViewModel() {
    private val _uiState = MutableStateFlow(ScanUiState())
    val uiState: StateFlow<ScanUiState> = _uiState.asStateFlow()

    private val scanCallback = object : Callback {
        override fun onProgress(tested: Long, healthy: Long, failed: Long, inFlight: Long, isPhase2: Boolean) {
            val current = _uiState.value
            var total = current.totalPhase2
            if (isPhase2 && total == 0 && tested.toInt() == 0) {
                total = inFlight.toInt()
            }
            _uiState.value = current.copy(
                tested = tested.toInt(),
                healthy = healthy.toInt(),
                failed = failed.toInt(),
                inFlight = inFlight.toInt(),
                isPhase2 = isPhase2,
                totalPhase2 = total
            )
        }

        override fun onResult(ip: String, port: Long, latencyMs: Long, loss: Double, colo: String, isHealthy: Boolean, isPhase2: Boolean, phase2Type: String, phase2Speed: Double, phase2Status: Boolean) {
            val res = IpResult(ip, port.toInt(), latencyMs.toInt(), loss, colo, isHealthy, isPhase2, phase2Type, phase2Speed, phase2Status)
            val newList = _uiState.value.results.toMutableList()
            newList.add(0, res)
            _uiState.value = _uiState.value.copy(results = newList)
        }

        override fun onFinished() {
            _uiState.value = _uiState.value.copy(isRunning = false)
        }

        override fun onError(err: String) {
            _uiState.value = _uiState.value.copy(isRunning = false, error = err)
        }
    }

    fun updateConfig(config: ScanConfig) {
        _uiState.value = _uiState.value.copy(config = config)
    }

    fun toggleScan() {
        if (Mobile.isRunning()) {
            Mobile.stopScan()
            _uiState.value = _uiState.value.copy(isRunning = false)
        } else {
            val jsonConfig = Json.encodeToString(_uiState.value.config)
            _uiState.value = _uiState.value.copy(
                isRunning = true,
                tested = 0,
                healthy = 0,
                failed = 0,
                inFlight = 0,
                totalPhase2 = 0,
                results = emptyList(),
                error = null
            )
            Mobile.startScan(jsonConfig, scanCallback)
        }
    }
    
    fun dismissError() {
        _uiState.value = _uiState.value.copy(error = null)
    }
}
