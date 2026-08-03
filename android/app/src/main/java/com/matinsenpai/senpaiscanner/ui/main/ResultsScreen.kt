package com.matinsenpai.senpaiscanner.ui.main

import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.outlined.Bolt
import androidx.compose.material.icons.outlined.CheckCircle
import androidx.compose.material.icons.outlined.ContentCopy
import androidx.compose.material.icons.outlined.Search
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.FilterChip
import androidx.compose.material3.FilterChipDefaults
import androidx.compose.material3.Icon
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.OutlinedTextFieldDefaults
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.derivedStateOf
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalClipboardManager
import androidx.compose.ui.text.AnnotatedString
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.matinsenpai.senpaiscanner.theme.SignalBorder
import com.matinsenpai.senpaiscanner.theme.SignalCyan
import com.matinsenpai.senpaiscanner.theme.SignalDanger
import com.matinsenpai.senpaiscanner.theme.SignalGreen
import com.matinsenpai.senpaiscanner.theme.SignalMuted
import com.matinsenpai.senpaiscanner.theme.SignalPanel
import com.matinsenpai.senpaiscanner.theme.SignalPanelRaised
import com.matinsenpai.senpaiscanner.theme.SignalText
import java.util.Locale

private enum class ResultSort { LATENCY, IP, COLO }

@Composable
fun ResultsScreen(uiState: ScanUiState, onSpeedTest: () -> Unit) {
    var query by rememberSaveable { mutableStateOf("") }
    var greenOnly by rememberSaveable { mutableStateOf(true) }
    var sort by rememberSaveable { mutableStateOf(ResultSort.LATENCY) }
    val clipboard = LocalClipboardManager.current
    val green = remember(uiState.results) { healthyEndpoints(uiState.results) }
    val top20 = remember(uiState.results) { healthyEndpoints(uiState.results, 20) }
    val phase1Rows by remember(uiState.results, query, greenOnly, sort) {
        derivedStateOf {
            uiState.results
                .filter { !it.isPhase2 && (!greenOnly || it.isHealthy) }
                .filter { query.isBlank() || it.ip.contains(query, true) || it.colo.contains(query, true) }
                .let { rows ->
                    when (sort) {
                        ResultSort.LATENCY -> rows.sortedBy { it.latencyMs }
                        ResultSort.IP -> rows.sortedBy { it.ip }
                        ResultSort.COLO -> rows.sortedBy { it.colo }
                    }
                }
        }
    }
    val speedRows = remember(uiState.results) {
        uiState.results.filter { it.isPhase2 }.sortedWith(
            compareByDescending<IpResult> { it.phase2Status }.thenByDescending { it.phase2Speed },
        )
    }

    LazyColumn(
        modifier = Modifier.fillMaxSize(),
        contentPadding = PaddingValues(14.dp, 14.dp, 14.dp, 28.dp),
        verticalArrangement = Arrangement.spacedBy(9.dp),
    ) {
        item {
            DeskPanel("LIVE RESULTS", "${green.size} green endpoints") {
                OutlinedTextField(
                    value = query,
                    onValueChange = { query = it },
                    modifier = Modifier.fillMaxWidth(),
                    singleLine = true,
                    leadingIcon = { Icon(Icons.Outlined.Search, contentDescription = null) },
                    placeholder = { Text("Search IP or colo") },
                    colors = OutlinedTextFieldDefaults.colors(
                        focusedTextColor = SignalText,
                        unfocusedTextColor = SignalText,
                        focusedBorderColor = SignalCyan,
                        unfocusedBorderColor = SignalBorder,
                        cursorColor = SignalCyan,
                        focusedLeadingIconColor = SignalCyan,
                        unfocusedLeadingIconColor = SignalMuted,
                        focusedPlaceholderColor = SignalMuted,
                        unfocusedPlaceholderColor = SignalMuted,
                    ),
                )
                Row(horizontalArrangement = Arrangement.spacedBy(7.dp)) {
                    FilterChip(
                        selected = greenOnly,
                        onClick = { greenOnly = true },
                        label = { Text("Green") },
                        colors = resultChipColors(),
                    )
                    FilterChip(
                        selected = !greenOnly,
                        onClick = { greenOnly = false },
                        label = { Text("All") },
                        colors = resultChipColors(),
                    )
                    Spacer(Modifier.weight(1f))
                    FilterChip(
                        selected = sort == ResultSort.LATENCY,
                        onClick = {
                            sort = when (sort) {
                                ResultSort.LATENCY -> ResultSort.IP
                                ResultSort.IP -> ResultSort.COLO
                                ResultSort.COLO -> ResultSort.LATENCY
                            }
                        },
                        label = { Text("Sort: ${sort.name.lowercase().replaceFirstChar { it.uppercase() }}") },
                        colors = resultChipColors(),
                    )
                }
                Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                    DeskAction(
                        label = "COPY GREEN",
                        enabled = green.isNotEmpty(),
                        onClick = { clipboard.setText(AnnotatedString(green.map { it.ip }.distinct().joinToString("\n"))) },
                        accent = SignalGreen,
                    )
                    DeskAction(
                        label = "COPY TOP 20",
                        enabled = top20.isNotEmpty(),
                        onClick = { clipboard.setText(AnnotatedString(top20.map { it.ip }.distinct().joinToString("\n"))) },
                    )
                }
                Button(
                    onClick = onSpeedTest,
                    enabled = !uiState.isRunning && green.isNotEmpty(),
                    modifier = Modifier.fillMaxWidth(),
                    shape = RoundedCornerShape(8.dp),
                    colors = ButtonDefaults.buttonColors(
                        containerColor = SignalCyan,
                        contentColor = Color(0xFF031116),
                        disabledContainerColor = SignalBorder,
                        disabledContentColor = SignalMuted,
                    ),
                ) {
                    Icon(Icons.Outlined.Bolt, contentDescription = null)
                    Text(
                        when {
                            uiState.isRunning -> "STOP THE SESSION FIRST"
                            green.isEmpty() -> "NO GREEN RESULTS YET"
                            else -> "SPEED TEST ${green.size} GREEN RESULTS"
                        },
                        fontWeight = FontWeight.Black,
                    )
                }
                if (uiState.isRunning) {
                    Text("Copy actions stay available while the scan is running.", color = SignalMuted, fontSize = 11.sp)
                }
            }
        }

        if (phase1Rows.isEmpty()) {
            item { EmptyResultPanel("Green endpoints will appear here in real time.") }
        } else {
            item { SectionLabel("PHASE 1  •  NETWORK PROBE") }
            items(phase1Rows, key = { "p1-${it.ip}:${it.port}" }) { result ->
                Phase1ResultCard(result) {
                    clipboard.setText(AnnotatedString("${result.ip}:${result.port}"))
                }
            }
        }

        if (speedRows.isNotEmpty()) {
            item {
                Spacer(Modifier.height(6.dp))
                SectionLabel("SPEED TEST  •  ${speedRows.count { it.phase2Status }} PASSED")
            }
            items(speedRows, key = { "p2-${it.ip}:${it.port}" }) { result ->
                SpeedResultCard(result) {
                    clipboard.setText(AnnotatedString("${result.ip}:${result.port}"))
                }
            }
        }
    }
}

@Composable
private fun Phase1ResultCard(result: IpResult, onCopy: () -> Unit) {
    Surface(color = SignalPanel, shape = RoundedCornerShape(8.dp), border = BorderStroke(1.dp, SignalBorder)) {
        Row(Modifier.fillMaxWidth().padding(12.dp), horizontalArrangement = Arrangement.spacedBy(10.dp)) {
            Icon(Icons.Outlined.CheckCircle, contentDescription = "Healthy", tint = SignalGreen)
            Column(Modifier.weight(1f)) {
                Text("${result.ip}:${result.port}", color = SignalText, fontFamily = FontFamily.Monospace, fontWeight = FontWeight.Bold)
                Text("${result.colo.ifBlank { "—" }}  •  loss ${String.format(Locale.US, "%.1f", result.loss)}%", color = SignalMuted, fontSize = 11.sp)
            }
            Column {
                Text("${result.latencyMs} ms", color = SignalGreen, fontWeight = FontWeight.Black)
                androidx.compose.material3.IconButton(onClick = onCopy) {
                    Icon(Icons.Outlined.ContentCopy, contentDescription = "Copy endpoint", tint = SignalCyan)
                }
            }
        }
    }
}

@Composable
private fun SpeedResultCard(result: IpResult, onCopy: () -> Unit) {
    val accent = if (result.phase2Status) SignalGreen else SignalDanger
    Surface(color = SignalPanel, shape = RoundedCornerShape(8.dp), border = BorderStroke(1.dp, accent.copy(alpha = .55f))) {
        Column(Modifier.fillMaxWidth().padding(12.dp), verticalArrangement = Arrangement.spacedBy(5.dp)) {
            Row {
                Text("${result.ip}:${result.port}", color = SignalText, fontFamily = FontFamily.Monospace, fontWeight = FontWeight.Bold, modifier = Modifier.weight(1f))
                Text(if (result.phase2Status) "PASSED" else "FAILED", color = accent, fontSize = 11.sp, fontWeight = FontWeight.Black)
            }
            Row {
                Text(result.phase2Type.ifBlank { "direct" }.uppercase(), color = SignalCyan, fontSize = 11.sp, modifier = Modifier.weight(1f))
                Text(formatSpeed(result.phase2Speed), color = accent, fontWeight = FontWeight.Bold)
                androidx.compose.material3.IconButton(onClick = onCopy) {
                    Icon(Icons.Outlined.ContentCopy, contentDescription = "Copy speed-tested endpoint", tint = SignalCyan)
                }
            }
            if (result.phase2UploadSpeed > 0) {
                Text("UPLOAD  ${formatSpeed(result.phase2UploadSpeed)}", color = SignalMuted, fontSize = 11.sp)
            }
        }
    }
}

@Composable
private fun SectionLabel(text: String) {
    Text(text, color = SignalCyan, fontSize = 10.sp, fontWeight = FontWeight.Black, letterSpacing = 1.sp)
}

@Composable
private fun EmptyResultPanel(text: String) {
    Column(
        Modifier.fillMaxWidth().background(SignalPanelRaised, RoundedCornerShape(8.dp)).padding(20.dp),
    ) {
        Text("LISTENING FOR SIGNAL", color = SignalCyan, fontWeight = FontWeight.Black)
        Text(text, color = SignalMuted, fontSize = 12.sp)
    }
}

@Composable
private fun resultChipColors() = FilterChipDefaults.filterChipColors(
    containerColor = SignalPanelRaised,
    labelColor = SignalMuted,
    selectedContainerColor = SignalCyan.copy(alpha = .14f),
    selectedLabelColor = SignalCyan,
)

private fun formatSpeed(bytesPerSecond: Double): String = when {
    bytesPerSecond >= 1024 * 1024 -> String.format(Locale.US, "%.2f MB/s", bytesPerSecond / (1024 * 1024))
    bytesPerSecond > 0 -> String.format(Locale.US, "%.0f KB/s", bytesPerSecond / 1024)
    else -> "—"
}
