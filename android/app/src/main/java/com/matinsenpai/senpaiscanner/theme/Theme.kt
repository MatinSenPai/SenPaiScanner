package com.matinsenpai.senpaiscanner.theme

import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.darkColorScheme
import androidx.compose.runtime.Composable

private val SignalDeskColors = darkColorScheme(
    primary = SignalCyan,
    secondary = SignalGreen,
    tertiary = SignalAmber,
    background = SignalBackground,
    surface = SignalPanel,
    surfaceVariant = SignalPanelRaised,
    outline = SignalBorder,
    onPrimary = SignalBackground,
    onSecondary = SignalBackground,
    onBackground = SignalText,
    onSurface = SignalText,
    onSurfaceVariant = SignalMuted,
    error = SignalDanger,
    onError = SignalBackground,
)

@Composable
fun SenPaiScannerTheme(content: @Composable () -> Unit) {
    MaterialTheme(
        colorScheme = SignalDeskColors,
        typography = Typography,
        content = content,
    )
}
