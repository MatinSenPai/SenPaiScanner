package com.ehsatn.magik.ui.main

import android.content.ClipData
import android.content.ClipboardManager
import android.content.Context
import android.net.Uri
import android.widget.Toast
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.Image
import androidx.compose.foundation.background
import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ArrowDropDown
import androidx.compose.material.icons.filled.Home
import androidx.compose.material.icons.filled.Info
import androidx.compose.material.icons.filled.ExitToApp
import androidx.compose.material.icons.filled.Settings
import androidx.compose.material.icons.filled.Check
import androidx.compose.material.icons.filled.Close
import androidx.compose.material.icons.filled.ContentCopy
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.platform.LocalUriHandler
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.style.TextDecoration
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.window.Dialog
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import java.io.File
import java.io.FileOutputStream
import com.ehsatn.magik.BuildConfig
import com.ehsatn.magik.R
import com.ehsatn.magik.theme.MagikDarkBackground
import com.ehsatn.magik.theme.MagikError
import com.ehsatn.magik.theme.MagikOrange
import com.ehsatn.magik.theme.MagikSuccess
import com.ehsatn.magik.theme.MagikDarkSurface
import com.ehsatn.magik.theme.MagikPrimary

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun AppUI(viewModel: MainViewModel = viewModel()) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    var selectedTab by remember { mutableStateOf(0) }
    var showInfoDialog by remember { mutableStateOf(false) }
    val context = LocalContext.current

    LaunchedEffect(uiState.error) {
        uiState.error?.let { err ->
            Toast.makeText(context, err, Toast.LENGTH_LONG).show()
            viewModel.dismissError()
        }
    }

    if (showInfoDialog) {
        InfoDialog(onDismiss = { showInfoDialog = false })
    }

    Scaffold(
        topBar = {
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .background(MaterialTheme.colorScheme.background)
                    .statusBarsPadding()
                    .padding(horizontal = 16.dp, vertical = 8.dp),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Row(
                    modifier = Modifier
                        .clip(RoundedCornerShape(8.dp))
                        .clickable { showInfoDialog = true }
                        .padding(vertical = 8.dp),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Image(
                        painter = painterResource(id = R.drawable.ic_launcher_foreground_raw),
                        contentDescription = stringResource(R.string.info_app_logo),
                        modifier = Modifier.size(32.dp)
                    )
                    Spacer(modifier = Modifier.width(10.dp))
                    Text(
                        text = stringResource(R.string.app_name),
                        color = MagikOrange,
                        fontWeight = FontWeight.Bold,
                        fontSize = 20.sp
                    )
                }
                IconButton(onClick = { showInfoDialog = true }) {
                    Icon(
                        imageVector = Icons.Filled.Info,
                        contentDescription = stringResource(R.string.title_info),
                        tint = MagikOrange
                    )
                }
            }
        },
        bottomBar = {
            NavigationBar(containerColor = MaterialTheme.colorScheme.surface) {
                NavigationBarItem(
                    icon = { Icon(Icons.Default.Home, contentDescription = "Home") },
                    label = { Text("Home") },
                    selected = selectedTab == 0,
                    onClick = { selectedTab = 0 },
                    colors = NavigationBarItemDefaults.colors(
                        selectedIconColor = MaterialTheme.colorScheme.background,
                        selectedTextColor = MagikOrange,
                        indicatorColor = MagikOrange,
                        unselectedIconColor = Color.Gray,
                        unselectedTextColor = Color.Gray
                    )
                )
                NavigationBarItem(
                    icon = { Icon(Icons.Default.Settings, contentDescription = "Settings") },
                    label = { Text("Settings") },
                    selected = selectedTab == 1,
                    onClick = { selectedTab = 1 },
                    colors = NavigationBarItemDefaults.colors(
                        selectedIconColor = MaterialTheme.colorScheme.background,
                        selectedTextColor = MagikOrange,
                        indicatorColor = MagikOrange,
                        unselectedIconColor = Color.Gray,
                        unselectedTextColor = Color.Gray
                    )
                )
            }
        },
        floatingActionButton = {
            if (selectedTab == 0) {
                FloatingActionButton(
                    onClick = { viewModel.toggleScan() },
                    containerColor = if (uiState.isRunning) MagikError else MagikOrange,
                    contentColor = Color.White
                ) {
                    Text(
                        text = if (uiState.isRunning) "STOP SCAN" else "START SCAN",
                        modifier = Modifier.padding(horizontal = 16.dp),
                        fontWeight = FontWeight.Bold
                    )
                }
            }
        },
        floatingActionButtonPosition = FabPosition.Center
    ) { innerPadding ->
        Box(modifier = Modifier.padding(innerPadding).fillMaxSize()) {
            if (selectedTab == 0) {
                HomeScreen(uiState, context)
            } else {
                SettingsScreen(uiState.config) { newConfig ->
                    viewModel.updateConfig(newConfig)
                }
            }
        }
    }
}

@Composable
fun HomeScreen(uiState: ScanUiState, context: Context) {
        Column(modifier = Modifier.fillMaxSize().padding(16.dp)) {
        // Stats Cards
        Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(6.dp)) {
            StatCard("Tested", uiState.tested.toString(), Modifier.weight(1f))
            StatCard("In-Flight", uiState.inFlight.toString(), Modifier.weight(1f))
        }
        Spacer(modifier = Modifier.height(4.dp))
        Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(6.dp)) {
            StatCard("Healthy", uiState.healthy.toString(), Modifier.weight(1f), MagikSuccess)
            StatCard("Failed", uiState.failed.toString(), Modifier.weight(1f), MagikError)
        }

        Spacer(modifier = Modifier.height(8.dp))

        // Discovered IPs Header
        Text("Discovered IPs", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)
        
        Spacer(modifier = Modifier.height(8.dp))
        
        // Copy Buttons Row
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            // Phase 1 Copy Button
            OutlinedButton(
                onClick = {
                    val phase1Ips = uiState.results.filter { !it.isPhase2 && it.isHealthy }.map { it.ip }.distinct().joinToString("\n")
                    if (phase1Ips.isNotEmpty()) {
                        val clipboard = context.getSystemService(Context.CLIPBOARD_SERVICE) as ClipboardManager
                        clipboard.setPrimaryClip(ClipData.newPlainText("MAGIK IPs", phase1Ips))
                        val count = uiState.results.count { !it.isPhase2 && it.isHealthy }
                        Toast.makeText(context, "Copied $count Phase 1 IPs", Toast.LENGTH_SHORT).show()
                    }
                },
                modifier = Modifier.weight(1f),
                colors = ButtonDefaults.outlinedButtonColors(
                    contentColor = MagikOrange
                ),
                border = BorderStroke(1.dp, MagikOrange)
            ) {
                Text(
                    text = "Copy",
                    fontSize = 12.sp,
                    color = MagikOrange
                )
                Spacer(modifier = Modifier.width(4.dp))
                Text("Phase 1", fontSize = 13.sp, fontWeight = FontWeight.Bold, color = MagikOrange)
            }
            
            // Phase 2 Copy Button (only visible when Phase 2 results exist)
            if (uiState.results.any { it.isPhase2 }) {
                OutlinedButton(
                    onClick = {
                        val phase2Ips = uiState.results.filter { it.isPhase2 && it.phase2Status }.map { it.ip }.distinct().joinToString("\n")
                        if (phase2Ips.isNotEmpty()) {
                            val clipboard = context.getSystemService(Context.CLIPBOARD_SERVICE) as ClipboardManager
                            clipboard.setPrimaryClip(ClipData.newPlainText("MAGIK Phase 2 IPs", phase2Ips))
                            val count = uiState.results.count { it.isPhase2 && it.phase2Status }
                            Toast.makeText(context, "Copied $count Phase 2 IPs", Toast.LENGTH_SHORT).show()
                        }
                    },
                    modifier = Modifier.weight(1f),
                    colors = ButtonDefaults.outlinedButtonColors(
                        contentColor = MagikPrimary
                    ),
                    border = BorderStroke(1.dp, MagikPrimary)
                ) {
                    Text(
                        text = "Copy",
                        fontSize = 12.sp,
                        color = MagikPrimary
                    )
                    Spacer(modifier = Modifier.width(4.dp))
                    Text("Phase 2", fontSize = 13.sp, fontWeight = FontWeight.Bold, color = MagikPrimary)
                }
            }
        }

        if (uiState.isPhase2) {
            val progress = if (uiState.totalPhase2 > 0) uiState.tested.toFloat() / uiState.totalPhase2 else 0f
            Column(modifier = Modifier.padding(bottom = 8.dp)) {
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Text(
                        text = "Xray Validating Candidates...",
                        color = MagikPrimary,
                        fontWeight = FontWeight.Bold
                    )
                    Text(
                        text = "${uiState.tested} / ${uiState.totalPhase2}",
                        color = MagikPrimary,
                        fontWeight = FontWeight.Bold,
                        fontSize = 14.sp
                    )
                }
                Spacer(modifier = Modifier.height(4.dp))
                LinearProgressIndicator(
                    progress = { progress },
                    modifier = Modifier.fillMaxWidth().height(6.dp),
                    color = MagikPrimary,
                    trackColor = MagikDarkSurface,
                )
                Spacer(modifier = Modifier.height(4.dp))
                Text(
                    text = "${(progress * 100).toInt()}%",
                    color = Color.Gray,
                    fontSize = 12.sp
                )
            }
        }

        LazyColumn(
            modifier = Modifier.fillMaxSize(),
            contentPadding = PaddingValues(bottom = 80.dp) // space for FAB
        ) {
            items(uiState.results) { res ->
                Card(
                    modifier = Modifier.fillMaxWidth().padding(vertical = 4.dp),
                    colors = CardDefaults.cardColors(containerColor = MagikDarkSurface)
                ) {
                    if (res.isPhase2) {
                        Column(modifier = Modifier.padding(12.dp).fillMaxWidth()) {
                            Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween, verticalAlignment = Alignment.CenterVertically) {
                                Row(verticalAlignment = Alignment.CenterVertically, modifier = Modifier.weight(1f)) {
                                    Text("${res.ip}:${res.port}", fontWeight = FontWeight.Bold, fontSize = 14.sp, maxLines = 1)
                                    Spacer(modifier = Modifier.width(4.dp))
                                    Icon(
                                        imageVector = Icons.Default.ContentCopy,
                                        contentDescription = "Copy IP",
                                        tint = MagikOrange,
                                        modifier = Modifier
                                            .size(16.dp)
                                            .clickable {
                                                val clipboard = context.getSystemService(Context.CLIPBOARD_SERVICE) as ClipboardManager
                                                clipboard.setPrimaryClip(ClipData.newPlainText("IP", res.ip))
                                                Toast.makeText(context, "IP copied: ${res.ip}", Toast.LENGTH_SHORT).show()
                                            }
                                    )
                                }
                                Icon(
                                    imageVector = if (res.phase2Status) Icons.Default.Check else Icons.Default.Close,
                                    contentDescription = if (res.phase2Status) "Passed" else "Failed",
                                    tint = if (res.phase2Status) MagikSuccess else MagikError,
                                    modifier = Modifier.size(20.dp)
                                )
                            }
                            Spacer(modifier = Modifier.height(4.dp))
                            Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween) {
                                Text("Type: ${res.phase2Type}", fontSize = 11.sp, color = Color.Gray)
                                val speedStr = if (res.phase2Speed > 1024*1024) String.format("%.2f MB/s", res.phase2Speed / (1024*1024)) else String.format("%.0f KB/s", res.phase2Speed / 1024)
                                Text("Speed: ${if (res.phase2Speed > 0) speedStr else "-"}", fontSize = 11.sp, color = Color.Gray)
                                Text("Latency: ${if (res.latencyMs > 0) "${res.latencyMs}ms" else "-"}", fontSize = 11.sp, color = Color.Gray)
                            }
                        }
                    } else {
                        Row(modifier = Modifier.padding(16.dp).fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween, verticalAlignment = Alignment.CenterVertically) {
                            Row(verticalAlignment = Alignment.CenterVertically) {
                                Column {
                                    Text(res.ip, fontWeight = FontWeight.Bold, fontSize = 16.sp)
                                    Text("Port: ${res.port} | Colo: ${res.colo}", fontSize = 12.sp, color = Color.Gray)
                                }
                                Spacer(modifier = Modifier.width(8.dp))
                                Icon(
                                    imageVector = Icons.Default.ContentCopy,
                                    contentDescription = "Copy IP",
                                    tint = MagikOrange,
                                    modifier = Modifier
                                        .size(18.dp)
                                        .clickable {
                                            val clipboard = context.getSystemService(Context.CLIPBOARD_SERVICE) as ClipboardManager
                                            clipboard.setPrimaryClip(ClipData.newPlainText("IP", res.ip))
                                            Toast.makeText(context, "IP copied: ${res.ip}", Toast.LENGTH_SHORT).show()
                                        }
                                )
                            }
                            Column(horizontalAlignment = Alignment.End) {
                                Text("${res.latencyMs} ms", color = if (res.isHealthy) MagikSuccess else MagikError, fontWeight = FontWeight.Bold)
                                Text("Loss: ${String.format("%.2f", res.loss)}%", fontSize = 12.sp, color = Color.Gray)
                            }
                        }
                    }
                }
            }
        }
    }
}

@Composable
fun StatCard(title: String, value: String, modifier: Modifier = Modifier, valueColor: Color = Color.White) {
    Card(
        modifier = modifier,
        colors = CardDefaults.cardColors(containerColor = MagikDarkSurface)
    ) {
        Column(modifier = Modifier.padding(horizontal = 12.dp, vertical = 8.dp)) {
            Text(title, color = Color.Gray, fontSize = 11.sp)
            Text(value, color = valueColor, fontSize = 20.sp, fontWeight = FontWeight.Bold)
        }
    }
}

@Composable
fun SettingsScreen(config: ScanConfig, onConfigChanged: (ScanConfig) -> Unit) {
    val context = LocalContext.current

    val filePickerLauncher = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.GetContent()
    ) { uri: Uri? ->
        if (uri != null) {
            try {
                val inputStream = context.contentResolver.openInputStream(uri)
                val tempFile = File(context.cacheDir, "ips.txt")
                val outputStream = FileOutputStream(tempFile)
                inputStream?.copyTo(outputStream)
                inputStream?.close()
                outputStream.close()
                onConfigChanged(config.copy(sourceFile = tempFile.absolutePath, sourceType = "From File"))
                Toast.makeText(context, "File selected", Toast.LENGTH_SHORT).show()
            } catch (e: Exception) {
                Toast.makeText(context, "Failed to load file", Toast.LENGTH_SHORT).show()
            }
        }
    }

    LazyColumn(modifier = Modifier.fillMaxSize().padding(16.dp)) {
        item {
            Text("Scanner Settings", style = MaterialTheme.typography.headlineSmall, color = MagikOrange, fontWeight = FontWeight.Bold)
            Spacer(modifier = Modifier.height(16.dp))
        }

        // 1. Source
        item {
            SettingSection("Source", "") {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    RadioButton(selected = config.sourceType == "Random", onClick = { onConfigChanged(config.copy(sourceType = "Random")) }, colors = RadioButtonDefaults.colors(selectedColor = MagikOrange))
                    Text("Random", modifier = Modifier.clickable { onConfigChanged(config.copy(sourceType = "Random")) })
                    Spacer(modifier = Modifier.width(16.dp))
                    RadioButton(selected = config.sourceType == "From File", onClick = { onConfigChanged(config.copy(sourceType = "From File")) }, colors = RadioButtonDefaults.colors(selectedColor = MagikOrange))
                    Text("From File", modifier = Modifier.clickable { onConfigChanged(config.copy(sourceType = "From File")) })
                }
                if (config.sourceType == "From File") {
                    Spacer(modifier = Modifier.height(8.dp))
                    Button(onClick = { filePickerLauncher.launch("text/plain") }, colors = ButtonDefaults.buttonColors(containerColor = MagikDarkSurface)) {
                        Text(if (config.sourceFile.isNotEmpty()) "File Selected" else "Import .txt File", color = MagikOrange)
                    }
                }
            }
        }

        // 2. Count
        item {
            SettingDropdown(
                label = "Count",
                description = "IPs to probe in Phase 1",
                options = listOf("1000", "5000", "20000", "Custom"),
                selectedOption = config.countType,
                customValue = config.customCount,
                onOptionSelected = { onConfigChanged(config.copy(countType = it)) },
                onCustomValueChanged = { onConfigChanged(config.copy(customCount = it)) },
                isNumericCustom = true
            )
        }

        // 3. Workers
        item {
            SettingDropdown(
                label = "Workers",
                description = "concurrent probes",
                options = listOf("50- default (restricted net)", "100 - balanced", "200 - fast (good connections)", "Custom"),
                selectedOption = config.workerType,
                customValue = config.customWorkers,
                onOptionSelected = { onConfigChanged(config.copy(workerType = it)) },
                onCustomValueChanged = { onConfigChanged(config.copy(customWorkers = it)) },
                isNumericCustom = true
            )
        }

        // 4. Timeout
        item {
            SettingDropdown(
                label = "Timeout",
                description = "per-probe deadline",
                options = listOf("2s - aggressive (fast net)", "3s- balanced", "5s - default (restricted net)", "Custom"),
                selectedOption = config.timeoutType,
                customValue = config.customTimeout,
                onOptionSelected = { onConfigChanged(config.copy(timeoutType = it)) },
                onCustomValueChanged = { onConfigChanged(config.copy(customTimeout = it)) },
                isNumericCustom = false // e.g., "10s"
            )
        }

        // 5. Ports
        item {
            SettingSection("Ports", "selecting multiple ports multiplies work") {
                val portOptions = listOf("Config", "443", "8443", "2053", "2083", "2087", "2096")
                Column {
                    portOptions.chunked(3).forEach { rowOptions ->
                        Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                            rowOptions.forEach { opt ->
                                Row(verticalAlignment = Alignment.CenterVertically) {
                                    if (opt == "Config") {
                                        RadioButton(
                                            selected = config.portType == "Config",
                                            onClick = { onConfigChanged(config.copy(portType = "Config")) },
                                            colors = RadioButtonDefaults.colors(selectedColor = MagikOrange)
                                        )
                                        Text("Config", modifier = Modifier.clickable { onConfigChanged(config.copy(portType = "Config")) })
                                    } else {
                                        Checkbox(
                                            checked = config.portType == "CustomPorts" && config.selectedPorts.contains(opt.toInt()),
                                            onCheckedChange = { checked ->
                                                val newSet = config.selectedPorts.toMutableSet()
                                                if (checked) newSet.add(opt.toInt()) else newSet.remove(opt.toInt())
                                                onConfigChanged(config.copy(portType = "CustomPorts", selectedPorts = newSet, configUrl = ""))
                                            },
                                            colors = CheckboxDefaults.colors(checkedColor = MagikOrange)
                                        )
                                        Text(opt)
                                    }
                                }
                            }
                        }
                    }
                    if (config.portType == "Config") {
                        Spacer(modifier = Modifier.height(8.dp))
                        OutlinedTextField(
                            value = config.configUrl,
                            onValueChange = { onConfigChanged(config.copy(configUrl = it)) },
                            label = { Text("Config URL (vless://...)") },
                            modifier = Modifier.fillMaxWidth()
                        )
                    }
                }
            }
        }

        // 6. Top N
        item {
            SettingDropdown(
                label = "Top N",
                description = "Phase 2 picks - used only when a config URL is entered",
                options = listOf("10", "25", "50", "100", "ALL", "Custom"),
                selectedOption = config.topNType,
                customValue = config.customTopN,
                onOptionSelected = { onConfigChanged(config.copy(topNType = it)) },
                onCustomValueChanged = { onConfigChanged(config.copy(customTopN = it)) },
                isNumericCustom = true
            )
        }

        item {
            Spacer(modifier = Modifier.height(80.dp))
        }
    }
}

@Composable
fun SettingSection(label: String, description: String, content: @Composable () -> Unit) {
    Column(modifier = Modifier.fillMaxWidth().padding(bottom = 16.dp)) {
        Text(label, fontWeight = FontWeight.Bold, fontSize = 16.sp)
        if (description.isNotEmpty()) {
            Text(description, fontSize = 12.sp, color = Color.Gray, modifier = Modifier.padding(bottom = 8.dp))
        } else {
            Spacer(modifier = Modifier.height(8.dp))
        }
        content()
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun SettingDropdown(
    label: String,
    description: String,
    options: List<String>,
    selectedOption: String,
    customValue: String,
    onOptionSelected: (String) -> Unit,
    onCustomValueChanged: (String) -> Unit,
    isNumericCustom: Boolean
) {
    var expanded by remember { mutableStateOf(false) }

    SettingSection(label, description) {
        ExposedDropdownMenuBox(
            expanded = expanded,
            onExpandedChange = { expanded = !expanded }
        ) {
            OutlinedTextField(
                value = selectedOption,
                onValueChange = {},
                readOnly = true,
                modifier = Modifier.menuAnchor().fillMaxWidth(),
                trailingIcon = { ExposedDropdownMenuDefaults.TrailingIcon(expanded = expanded) }
            )
            ExposedDropdownMenu(
                expanded = expanded,
                onDismissRequest = { expanded = false }
            ) {
                options.forEach { selectionOption ->
                    DropdownMenuItem(
                        text = { Text(selectionOption) },
                        onClick = {
                            onOptionSelected(selectionOption)
                            expanded = false
                        }
                    )
                }
            }
        }
        
        if (selectedOption == "Custom") {
            Spacer(modifier = Modifier.height(8.dp))
            OutlinedTextField(
                value = customValue,
                onValueChange = onCustomValueChanged,
                label = { Text("Enter Custom Value") },
                keyboardOptions = if (isNumericCustom) KeyboardOptions(keyboardType = KeyboardType.Number) else KeyboardOptions.Default,
                modifier = Modifier.fillMaxWidth()
            )
        }
    }
}

@Composable
fun InfoDialog(onDismiss: () -> Unit) {
    val uriHandler = LocalUriHandler.current

    Dialog(onDismissRequest = onDismiss) {
        Card(
            shape = RoundedCornerShape(20.dp),
            colors = CardDefaults.cardColors(containerColor = MagikDarkSurface),
            modifier = Modifier.fillMaxWidth()
        ) {
            Column(
                modifier = Modifier
                    .verticalScroll(rememberScrollState())
                    .padding(20.dp),
                verticalArrangement = Arrangement.spacedBy(16.dp)
            ) {
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Text(
                        text = stringResource(R.string.info_app_name_title),
                        style = MaterialTheme.typography.titleLarge,
                        fontWeight = FontWeight.Bold,
                        color = MagikOrange
                    )
                    TextButton(onClick = onDismiss) {
                        Text("X", color = Color.Gray, fontWeight = FontWeight.Bold)
                    }
                }

                Card(
                    shape = RoundedCornerShape(16.dp),
                    colors = CardDefaults.cardColors(containerColor = MagikDarkSurface)
                ) {
                    Box(
                        modifier = Modifier
                            .fillMaxWidth()
                            .background(
                                brush = Brush.linearGradient(
                                    colors = listOf(
                                        MagikOrange.copy(alpha = 0.15f),
                                        MagikOrange.copy(alpha = 0.05f)
                                    )
                                )
                            )
                            .padding(16.dp)
                    ) {
                        Row(verticalAlignment = Alignment.CenterVertically) {
                            Box(
                                modifier = Modifier
                                    .size(64.dp)
                                    .background(Color.White.copy(alpha = 0.1f), RoundedCornerShape(16.dp))
                                    .padding(4.dp)
                            ) {
                                Image(
                                    painter = painterResource(id = R.drawable.ic_launcher_foreground_raw),
                                    contentDescription = stringResource(R.string.info_app_logo),
                                    modifier = Modifier
                                        .fillMaxSize()
                                        .clip(RoundedCornerShape(12.dp))
                                )
                            }
                            Spacer(modifier = Modifier.width(14.dp))
                            Column {
                                Text(
                                    text = stringResource(R.string.info_app_name_title),
                                    style = MaterialTheme.typography.titleMedium,
                                    fontWeight = FontWeight.Bold
                                )
                                Text(
                                    text = stringResource(R.string.info_overview_subtitle),
                                    style = MaterialTheme.typography.bodySmall,
                                    color = Color.Gray
                                )
                            }
                        }
                    }
                }

                Card(colors = CardDefaults.cardColors(containerColor = Color(0xFF2A2A2A))) {
                    Column(
                        modifier = Modifier.padding(14.dp),
                        verticalArrangement = Arrangement.spacedBy(6.dp)
                    ) {
                        Text("Description", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)
                        Text(
                            text = stringResource(R.string.info_description),
                            style = MaterialTheme.typography.bodyMedium,
                            color = Color.Gray
                        )
                    }
                }

                Card(colors = CardDefaults.cardColors(containerColor = Color(0xFF2A2A2A))) {
                    Column(
                        modifier = Modifier.padding(14.dp),
                        verticalArrangement = Arrangement.spacedBy(6.dp)
                    ) {
                        val githubLink = stringResource(R.string.project_main_github)
                        val telegramLink = stringResource(R.string.project_main_telegram)
                        Text(stringResource(R.string.info_project_links), style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)
                        InfoLinkRow(
                            title = stringResource(R.string.info_main_github),
                            link = githubLink,
                            onOpen = { uriHandler.openUri("https://$githubLink") }
                        )
                        InfoLinkRow(
                            title = stringResource(R.string.info_main_telegram),
                            link = telegramLink,
                            onOpen = { uriHandler.openUri("https://$telegramLink") }
                        )
                    }
                }

                Card(colors = CardDefaults.cardColors(containerColor = Color(0xFF2A2A2A))) {
                    Column(
                        modifier = Modifier.padding(14.dp),
                        verticalArrangement = Arrangement.spacedBy(8.dp)
                    ) {
                        Text(stringResource(R.string.info_version_info), style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)
                        InfoValueRow(label = stringResource(R.string.info_app_version), value = BuildConfig.VERSION_NAME)
                    }
                }
            }
        }
    }
}

@Composable
private fun InfoLinkRow(title: String, link: String, onOpen: () -> Unit) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .clip(RoundedCornerShape(10.dp))
            .clickable(onClick = onOpen)
            .padding(horizontal = 8.dp, vertical = 10.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.SpaceBetween
    ) {
        Column(modifier = Modifier.weight(1f)) {
            Text(text = title, style = MaterialTheme.typography.labelLarge)
            Spacer(modifier = Modifier.height(2.dp))
            Text(
                text = link,
                style = MaterialTheme.typography.bodySmall.copy(textDecoration = TextDecoration.Underline),
                color = MagikOrange,
                maxLines = 2
            )
        }
        Spacer(modifier = Modifier.width(10.dp))
        Icon(
            imageVector = Icons.Filled.ExitToApp,
            contentDescription = stringResource(R.string.info_open_link),
            tint = MagikOrange
        )
    }
}

@Composable
private fun InfoValueRow(label: String, value: String) {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .clip(RoundedCornerShape(12.dp))
            .background(Color(0xFF333333))
            .padding(horizontal = 12.dp, vertical = 10.dp)
    ) {
        Text(
            text = label,
            style = MaterialTheme.typography.labelMedium,
            color = Color.Gray
        )
        Spacer(modifier = Modifier.height(4.dp))
        Text(
            text = value,
            style = MaterialTheme.typography.bodyLarge,
            fontWeight = FontWeight.Medium
        )
    }
}
