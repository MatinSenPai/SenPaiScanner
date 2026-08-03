package com.matinsenpai.senpaiscanner.ui.main

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.ExperimentalLayoutApi
import androidx.compose.foundation.layout.FlowRow
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.outlined.PlayArrow
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.FilterChip
import androidx.compose.material3.FilterChipDefaults
import androidx.compose.material3.Icon
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.OutlinedTextFieldDefaults
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.dp
import com.matinsenpai.senpaiscanner.theme.SignalBorder
import com.matinsenpai.senpaiscanner.theme.SignalCyan
import com.matinsenpai.senpaiscanner.theme.SignalMuted
import com.matinsenpai.senpaiscanner.theme.SignalPanelRaised
import com.matinsenpai.senpaiscanner.theme.SignalText

private val scanPorts = listOf(443, 80, 8443, 2053, 2083, 2087, 2096)

@OptIn(ExperimentalLayoutApi::class)
@Composable
fun ScanScreen(
    config: ScanConfig,
    running: Boolean,
    onConfigChange: (ScanConfig) -> Unit,
    onStart: () -> Unit,
) {
    LazyColumn(
        modifier = Modifier.fillMaxSize(),
        contentPadding = PaddingValues(14.dp, 14.dp, 14.dp, 28.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        item {
            DeskPanel("01 / ESSENTIALS", "Choose the scan envelope") {
                FieldLabel("IP count")
                ChoiceChips(
                    choices = listOf("1K" to "1000", "5K" to "5000", "20K" to "20000"),
                    selected = config.countType,
                    onSelected = { onConfigChange(config.copy(countType = it, customCount = "")) },
                )

                FieldLabel("Parallel workers")
                ChoiceChips(
                    choices = listOf("50" to "50", "100" to "100", "200" to "200"),
                    selected = config.workerType.substringBefore("-"),
                    onSelected = { onConfigChange(config.copy(workerType = it, customWorkers = "")) },
                )

                FieldLabel("Probe timeout")
                ChoiceChips(
                    choices = listOf("2 sec" to "2s", "3 sec" to "3s", "5 sec" to "5s"),
                    selected = config.timeoutType.substringBefore(" "),
                    onSelected = { onConfigChange(config.copy(timeoutType = it, customTimeout = "")) },
                )

                FieldLabel("Ports")
                FlowRow(horizontalArrangement = Arrangement.spacedBy(7.dp), verticalArrangement = Arrangement.spacedBy(7.dp)) {
                    scanPorts.forEach { port ->
                        val selected = port in config.selectedPorts
                        FilterChip(
                            selected = selected,
                            onClick = {
                                val updated = config.selectedPorts.toMutableSet().apply {
                                    if (selected) remove(port) else add(port)
                                }
                                onConfigChange(config.copy(selectedPorts = updated.ifEmpty { mutableSetOf(443) }))
                            },
                            label = { Text(port.toString()) },
                            colors = FilterChipDefaults.filterChipColors(
                                containerColor = SignalPanelRaised,
                                labelColor = SignalMuted,
                                selectedContainerColor = SignalCyan.copy(alpha = .14f),
                                selectedLabelColor = SignalCyan,
                            ),
                            border = FilterChipDefaults.filterChipBorder(
                                enabled = true,
                                selected = selected,
                                borderColor = SignalBorder,
                                selectedBorderColor = SignalCyan,
                            ),
                        )
                    }
                }

                ToggleSetting(
                    title = "Neighbor scan",
                    supporting = "Optional. Expand around green IPs; off keeps resource use predictable.",
                    checked = config.neighborScan,
                    onCheckedChange = { onConfigChange(config.copy(neighborScan = it)) },
                )
                ToggleSetting(
                    title = "Require WebSocket",
                    supporting = "Only accept endpoints that complete the WebSocket probe.",
                    checked = config.requireWebSocket,
                    onCheckedChange = { onConfigChange(config.copy(requireWebSocket = it)) },
                )
            }
        }

        item {
            DeskPanel("02 / TUNNEL", "Optional proxy validation") {
                DeskTextField(
                    value = config.configUrl,
                    onValueChange = { onConfigChange(config.copy(configUrl = it.trim())) },
                    label = "VLESS, Trojan, or VMess URL",
                    placeholder = "Leave empty for direct speed testing",
                    minLines = 2,
                )

                FieldLabel("Automatic validation after scan")
                ChoiceChips(
                    choices = listOf("Top 20" to "20", "Top 50" to "50", "All" to "ALL"),
                    selected = config.topNType,
                    onSelected = { onConfigChange(config.copy(topNType = it)) },
                )

                DeskTextField(
                    value = if (config.minSpeed == 0.0) "" else config.minSpeed.toString(),
                    onValueChange = { onConfigChange(config.copy(minSpeed = it.toDoubleOrNull() ?: 0.0)) },
                    label = "Minimum speed (Mbps)",
                    placeholder = "0 = no threshold",
                    keyboardType = KeyboardType.Decimal,
                )

                FieldLabel("Download sample")
                ChoiceChips(
                    choices = listOf("512 KB" to 524_288L, "1 MB" to 1_048_576L, "5 MB" to 5_242_880L),
                    selected = config.speedSize,
                    onSelected = { onConfigChange(config.copy(speedSize = it)) },
                )

                ToggleSetting(
                    title = "Upload sample",
                    supporting = "Adds an upload measurement to tunnel validation.",
                    checked = config.uploadTest,
                    onCheckedChange = { onConfigChange(config.copy(uploadTest = it)) },
                )
            }
        }

        item {
            Button(
                onClick = onStart,
                enabled = !running,
                modifier = Modifier.fillMaxWidth(),
                shape = RoundedCornerShape(8.dp),
                colors = ButtonDefaults.buttonColors(
                    containerColor = SignalCyan,
                    contentColor = Color(0xFF031116),
                    disabledContainerColor = SignalBorder,
                    disabledContentColor = SignalMuted,
                ),
                contentPadding = PaddingValues(vertical = 15.dp),
            ) {
                Icon(Icons.Outlined.PlayArrow, contentDescription = null)
                Text(if (running) "SESSION IN PROGRESS" else "START SCAN", fontWeight = FontWeight.Black)
            }
        }
    }
}

@Composable
private fun FieldLabel(text: String) {
    Text(text, color = SignalText, fontWeight = FontWeight.SemiBold)
}

@Composable
fun DeskTextField(
    value: String,
    onValueChange: (String) -> Unit,
    label: String,
    placeholder: String = "",
    keyboardType: KeyboardType = KeyboardType.Text,
    minLines: Int = 1,
) {
    OutlinedTextField(
        value = value,
        onValueChange = onValueChange,
        label = { Text(label) },
        placeholder = { Text(placeholder) },
        modifier = Modifier.fillMaxWidth(),
        minLines = minLines,
        maxLines = if (minLines > 1) 4 else 1,
        keyboardOptions = KeyboardOptions(keyboardType = keyboardType),
        colors = OutlinedTextFieldDefaults.colors(
            focusedTextColor = SignalText,
            unfocusedTextColor = SignalText,
            focusedBorderColor = SignalCyan,
            unfocusedBorderColor = SignalBorder,
            focusedLabelColor = SignalCyan,
            unfocusedLabelColor = SignalMuted,
            cursorColor = SignalCyan,
            focusedPlaceholderColor = SignalMuted,
            unfocusedPlaceholderColor = SignalMuted,
        ),
    )
}
