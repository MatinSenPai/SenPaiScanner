package com.matinsenpai.senpaiscanner.theme

import androidx.compose.ui.graphics.Color

// Black, white, red - same palette as the desktop app. The Signal* names are kept so the rest of the UI code is untouched:
// SignalCyan is now the red accent, SignalGreen the white "ok" colour, SignalAmber the grey "warning" colour.
val SignalBackground = Color(0xFF0A0A0A)
val SignalPanel = Color(0xFF111111)
val SignalPanelRaised = Color(0xFF161616)
val SignalBorder = Color(0xFF2B2B2B)
val SignalGrid = Color(0xFF1E1E1E)
val SignalCyan = Color(0xFFEE2B38)
val SignalGreen = Color(0xFFF4F4F1)
val SignalAmber = Color(0xFF8D8D88)
val SignalDanger = Color(0xFFEE2B38)
val SignalText = Color(0xFFF4F4F1)
val SignalMuted = Color(0xFF8D8D88)

// Compatibility aliases for the small legacy sample screen still in-tree.
val SenPaiOrange = SignalCyan
val SenPaiDarkBackground = SignalBackground
val SenPaiDarkSurface = SignalPanel
val SenPaiSuccess = SignalGreen
val SenPaiError = SignalDanger
val SenPaiTextPrimary = SignalText
val SenPaiTextSecondary = SignalMuted
val SenPaiPrimary = SignalCyan
