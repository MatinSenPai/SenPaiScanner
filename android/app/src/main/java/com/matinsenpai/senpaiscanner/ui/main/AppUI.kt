package com.matinsenpai.senpaiscanner.ui.main

import androidx.compose.foundation.Canvas
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.outlined.FileDownload
import androidx.compose.material.icons.outlined.Radar
import androidx.compose.material.icons.outlined.TableRows
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Icon
import androidx.compose.material3.LinearProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.NavigationBar
import androidx.compose.material3.NavigationBarItem
import androidx.compose.material3.NavigationBarItemDefaults
import androidx.compose.material3.Scaffold
import androidx.compose.material3.SnackbarHost
import androidx.compose.material3.SnackbarHostState
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import com.matinsenpai.senpaiscanner.BuildConfig
import com.matinsenpai.senpaiscanner.R
import com.matinsenpai.senpaiscanner.theme.SignalCyan
import com.matinsenpai.senpaiscanner.theme.SignalDanger
import com.matinsenpai.senpaiscanner.theme.SignalGreen
import com.matinsenpai.senpaiscanner.theme.SignalGrid
import com.matinsenpai.senpaiscanner.theme.SignalMuted
import com.matinsenpai.senpaiscanner.theme.SignalPanel
import com.matinsenpai.senpaiscanner.theme.SignalText

enum class DeskTab(val label: String) { SCAN("Scan"), RESULTS("Results"), EXPORT("Export") }

@Composable
fun AppUI(viewModel: MainViewModel = viewModel()) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    SignalDeskApp(
        uiState = uiState,
        onConfigChange = viewModel::updateConfig,
        onStart = viewModel::startScan,
        onStop = viewModel::stop,
        onSpeedTest = viewModel::startSpeedTest,
        onGenerateExports = viewModel::generateExports,
        onDismissError = viewModel::dismissError,
    )
}

@Composable
fun SignalDeskApp(
    uiState: ScanUiState,
    onConfigChange: (ScanConfig) -> Unit,
    onStart: () -> Unit,
    onStop: () -> Unit,
    onSpeedTest: () -> Unit,
    onGenerateExports: () -> Unit,
    onDismissError: () -> Unit,
) {
    var selectedTab by rememberSaveable { mutableStateOf(DeskTab.SCAN) }
    val snackbar = remember { SnackbarHostState() }

    LaunchedEffect(uiState.error) {
        uiState.error?.let {
            snackbar.showSnackbar(it)
            onDismissError()
        }
    }

    Scaffold(
        containerColor = Color.Transparent,
        snackbarHost = { SnackbarHost(snackbar) },
        topBar = { BrandHeader() },
        bottomBar = {
            NavigationBar(containerColor = SignalPanel, tonalElevation = 0.dp) {
                DeskTab.entries.forEach { tab ->
                    val icon = when (tab) {
                        DeskTab.SCAN -> Icons.Outlined.Radar
                        DeskTab.RESULTS -> Icons.Outlined.TableRows
                        DeskTab.EXPORT -> Icons.Outlined.FileDownload
                    }
                    NavigationBarItem(
                        selected = selectedTab == tab,
                        onClick = { selectedTab = tab },
                        icon = { Icon(icon, contentDescription = null) },
                        label = { Text(tab.label) },
                        colors = NavigationBarItemDefaults.colors(
                            selectedIconColor = SignalCyan,
                            selectedTextColor = SignalCyan,
                            indicatorColor = SignalCyan.copy(alpha = .12f),
                            unselectedIconColor = SignalMuted,
                            unselectedTextColor = SignalMuted,
                        ),
                    )
                }
            }
        },
    ) { innerPadding ->
        SignalBackground(Modifier.fillMaxSize().padding(innerPadding)) {
            Column(Modifier.fillMaxSize()) {
                SessionRail(uiState = uiState, onStop = onStop)
                Box(Modifier.fillMaxSize()) {
                    when (selectedTab) {
                        DeskTab.SCAN -> ScanScreen(
                            config = uiState.config,
                            running = uiState.isRunning,
                            onConfigChange = onConfigChange,
                            onStart = {
                                selectedTab = DeskTab.RESULTS
                                onStart()
                            },
                        )
                        DeskTab.RESULTS -> ResultsScreen(
                            uiState = uiState,
                            onSpeedTest = onSpeedTest,
                        )
                        DeskTab.EXPORT -> ExportScreen(
                            uiState = uiState,
                            onGenerateExports = onGenerateExports,
                        )
                    }
                }
            }
        }
    }
}

@Composable
private fun BrandHeader() {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .background(SignalPanel)
            .statusBarsPadding()
            .padding(horizontal = 16.dp, vertical = 10.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        androidx.compose.foundation.Image(
            painter = painterResource(R.drawable.senpai_logo),
            contentDescription = "SenPai Scanner logo",
            modifier = Modifier.size(42.dp),
            contentScale = ContentScale.Fit,
        )
        Column(Modifier.padding(start = 10.dp).weight(1f)) {
            Text("SENPAI SCANNER", color = SignalText, fontWeight = FontWeight.Black, letterSpacing = 1.2.sp)
            Text("SIGNAL DESK  •  v${BuildConfig.VERSION_NAME}", color = SignalCyan, fontSize = 11.sp)
        }
        Text("ANDROID", color = SignalMuted, fontSize = 11.sp, fontWeight = FontWeight.Bold)
    }
}

@Composable
private fun SessionRail(uiState: ScanUiState, onStop: () -> Unit) {
    val status = when (uiState.phase) {
        SessionPhase.IDLE -> if (uiState.phase1Tested > 0) "READY" else "IDLE"
        SessionPhase.SCANNING -> "SCANNING"
        SessionPhase.STOPPING -> "STOPPING"
        SessionPhase.SPEED_TESTING -> "SPEED TEST"
    }
    val statusColor = when (uiState.phase) {
        SessionPhase.IDLE -> SignalGreen
        SessionPhase.STOPPING -> SignalDanger
        else -> SignalCyan
    }
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .background(SignalPanel.copy(alpha = .96f))
            .padding(horizontal = 14.dp, vertical = 10.dp)
            .semantics { contentDescription = "Session status $status" },
    ) {
        Row(verticalAlignment = Alignment.CenterVertically) {
            Text(status, color = statusColor, fontSize = 12.sp, fontWeight = FontWeight.Black, letterSpacing = 1.sp)
            Text(
                text = when {
                    uiState.isMetaLoading -> "  •  DETECTING ISP"
                    uiState.isp.isNotBlank() -> "  •  ${uiState.isp}"
                    else -> ""
                },
                color = SignalMuted,
                fontSize = 10.sp,
                maxLines = 1,
                modifier = Modifier.padding(start = 6.dp).weight(1f),
            )
            if (uiState.isRunning) {
                Button(
                    onClick = onStop,
                    enabled = uiState.phase != SessionPhase.STOPPING,
                    colors = ButtonDefaults.buttonColors(containerColor = SignalDanger),
                    modifier = Modifier.height(40.dp),
                ) { Text(if (uiState.phase == SessionPhase.STOPPING) "STOPPING…" else "STOP") }
            }
        }
        Spacer(Modifier.height(8.dp))
        Row(horizontalArrangement = Arrangement.spacedBy(8.dp), modifier = Modifier.fillMaxWidth()) {
            RailMetric("TESTED", uiState.phase1Tested.toString(), Modifier.weight(1f))
            RailMetric("GREEN", uiState.phase1Healthy.toString(), Modifier.weight(1f), SignalGreen)
            RailMetric("FAILED", uiState.phase1Failed.toString(), Modifier.weight(1f), SignalDanger)
            RailMetric("SPEED", "${uiState.speedDone}/${uiState.speedTotal}", Modifier.weight(1f), SignalCyan)
        }
        if (uiState.isRunning) {
            Spacer(Modifier.height(8.dp))
            if (uiState.phase == SessionPhase.SPEED_TESTING && uiState.speedTotal > 0) {
                LinearProgressIndicator(
                    progress = { uiState.speedDone.toFloat() / uiState.speedTotal },
                    modifier = Modifier.fillMaxWidth().height(3.dp),
                    color = SignalCyan,
                    trackColor = SignalGrid,
                )
            } else {
                LinearProgressIndicator(
                    modifier = Modifier.fillMaxWidth().height(3.dp),
                    color = SignalCyan,
                    trackColor = SignalGrid,
                )
            }
        }
    }
}

@Composable
private fun RailMetric(label: String, value: String, modifier: Modifier, valueColor: Color = SignalText) {
    Column(modifier) {
        Text(label, color = SignalMuted, fontSize = 9.sp, fontWeight = FontWeight.Bold)
        Text(value, color = valueColor, fontSize = 17.sp, fontWeight = FontWeight.Black)
    }
}

@Composable
private fun SignalBackground(modifier: Modifier = Modifier, content: @Composable () -> Unit) {
    Box(modifier.background(MaterialTheme.colorScheme.background)) {
        Canvas(Modifier.fillMaxSize()) {
            val step = 28.dp.toPx()
            var x = 0f
            while (x < size.width) {
                drawLine(SignalGrid.copy(alpha = .22f), Offset(x, 0f), Offset(x, size.height), 1f)
                x += step
            }
            var y = 0f
            while (y < size.height) {
                drawLine(SignalGrid.copy(alpha = .22f), Offset(0f, y), Offset(size.width, y), 1f)
                y += step
            }
        }
        content()
    }
}
