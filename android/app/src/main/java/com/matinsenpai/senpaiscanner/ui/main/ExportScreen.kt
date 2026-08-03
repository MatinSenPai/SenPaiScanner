package com.matinsenpai.senpaiscanner.ui.main

import android.content.Intent
import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.outlined.AutoAwesome
import androidx.compose.material.icons.outlined.ContentCopy
import androidx.compose.material.icons.outlined.Share
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Icon
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalClipboardManager
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.AnnotatedString
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.matinsenpai.senpaiscanner.theme.SignalBorder
import com.matinsenpai.senpaiscanner.theme.SignalCyan
import com.matinsenpai.senpaiscanner.theme.SignalGreen
import com.matinsenpai.senpaiscanner.theme.SignalMuted
import com.matinsenpai.senpaiscanner.theme.SignalPanel
import com.matinsenpai.senpaiscanner.theme.SignalText

@Composable
fun ExportScreen(uiState: ScanUiState, onGenerateExports: () -> Unit) {
    val context = LocalContext.current
    val clipboard = LocalClipboardManager.current
    val greenText = remember(uiState.results) {
        healthyEndpoints(uiState.results).joinToString("\n") { "${it.ip}:${it.port}" }
    }
    val top20Text = remember(uiState.results) {
        healthyEndpoints(uiState.results, 20).joinToString("\n") { "${it.ip}:${it.port}" }
    }
    val passedCount = uiState.speedResults.count { it.phase2Status }

    LazyColumn(
        modifier = Modifier.fillMaxSize(),
        contentPadding = PaddingValues(14.dp, 14.dp, 14.dp, 28.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        item {
            DeskPanel("RAW ENDPOINTS", "Copy or share live scan data") {
                Text(
                    "These actions use the complete in-memory result set, even while scanning.",
                    color = SignalMuted,
                    fontSize = 12.sp,
                )
                Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                    DeskAction(
                        label = "ALL GREEN",
                        enabled = greenText.isNotBlank(),
                        onClick = { clipboard.setText(AnnotatedString(greenText)) },
                        accent = SignalGreen,
                    )
                    DeskAction(
                        label = "TOP 20",
                        enabled = top20Text.isNotBlank(),
                        onClick = { clipboard.setText(AnnotatedString(top20Text)) },
                    )
                }
                OutlinedButton(
                    onClick = { shareText(context, "SenPaiScanner green endpoints", greenText) },
                    enabled = greenText.isNotBlank(),
                    modifier = Modifier.fillMaxWidth(),
                    border = BorderStroke(1.dp, SignalBorder),
                    shape = RoundedCornerShape(8.dp),
                ) {
                    Icon(Icons.Outlined.Share, contentDescription = null)
                    Text("SHARE GREEN ENDPOINTS")
                }
            }
        }

        item {
            DeskPanel("CLIENT CONFIGS", "$passedCount speed-tested endpoints") {
                Text(
                    "Generate a subscription, sing-box JSON and Clash YAML from passing tunnel results.",
                    color = SignalMuted,
                    fontSize = 12.sp,
                )
                Button(
                    onClick = onGenerateExports,
                    enabled = !uiState.isRunning && passedCount > 0 && uiState.config.configUrl.isNotBlank(),
                    modifier = Modifier.fillMaxWidth(),
                    colors = ButtonDefaults.buttonColors(
                        containerColor = SignalCyan,
                        contentColor = Color(0xFF031116),
                        disabledContainerColor = SignalBorder,
                        disabledContentColor = SignalMuted,
                    ),
                    shape = RoundedCornerShape(8.dp),
                ) {
                    Icon(Icons.Outlined.AutoAwesome, contentDescription = null)
                    Text("GENERATE CLIENT CONFIGS", fontWeight = FontWeight.Black)
                }
                if (uiState.config.configUrl.isBlank()) {
                    Text("Add a proxy URL in Scan to enable client exports.", color = SignalMuted, fontSize = 11.sp)
                } else if (passedCount == 0) {
                    Text("Run Speed Test first; only passing endpoints are exported.", color = SignalMuted, fontSize = 11.sp)
                }
            }
        }

        uiState.exportBundle?.let { bundle ->
            item {
                ExportOutputCard(
                    title = "SUBSCRIPTION",
                    subtitle = "${bundle.count} share URLs",
                    preview = bundle.subscription,
                    onCopy = { clipboard.setText(AnnotatedString(bundle.subscription)) },
                    onShare = { shareText(context, "SenPaiScanner subscription", bundle.subscription) },
                )
            }
            item {
                ExportOutputCard(
                    title = "SING-BOX JSON",
                    subtitle = "Ready for import",
                    preview = bundle.singBox,
                    onCopy = { clipboard.setText(AnnotatedString(bundle.singBox)) },
                    onShare = { shareText(context, "SenPaiScanner sing-box", bundle.singBox) },
                )
            }
            item {
                ExportOutputCard(
                    title = "CLASH YAML",
                    subtitle = "Proxy list",
                    preview = bundle.clash,
                    onCopy = { clipboard.setText(AnnotatedString(bundle.clash)) },
                    onShare = { shareText(context, "SenPaiScanner Clash", bundle.clash) },
                )
            }
        }
    }
}

@Composable
private fun ExportOutputCard(
    title: String,
    subtitle: String,
    preview: String,
    onCopy: () -> Unit,
    onShare: () -> Unit,
) {
    Surface(color = SignalPanel, shape = RoundedCornerShape(8.dp), border = BorderStroke(1.dp, SignalBorder)) {
        Column(Modifier.fillMaxWidth().padding(14.dp), verticalArrangement = Arrangement.spacedBy(9.dp)) {
            Row {
                Column(Modifier.weight(1f)) {
                    Text(title, color = SignalCyan, fontWeight = FontWeight.Black, letterSpacing = 1.sp)
                    Text(subtitle, color = SignalMuted, fontSize = 11.sp)
                }
                androidx.compose.material3.IconButton(onClick = onCopy) {
                    Icon(Icons.Outlined.ContentCopy, contentDescription = "Copy $title", tint = SignalCyan)
                }
                androidx.compose.material3.IconButton(onClick = onShare) {
                    Icon(Icons.Outlined.Share, contentDescription = "Share $title", tint = SignalCyan)
                }
            }
            Text(
                preview.lineSequence().take(5).joinToString("\n"),
                color = SignalText,
                fontFamily = FontFamily.Monospace,
                fontSize = 10.sp,
                lineHeight = 14.sp,
                maxLines = 5,
            )
        }
    }
}

private fun shareText(context: android.content.Context, title: String, text: String) {
    val intent = Intent(Intent.ACTION_SEND).apply {
        type = "text/plain"
        putExtra(Intent.EXTRA_SUBJECT, title)
        putExtra(Intent.EXTRA_TEXT, text)
    }
    context.startActivity(Intent.createChooser(intent, title))
}
