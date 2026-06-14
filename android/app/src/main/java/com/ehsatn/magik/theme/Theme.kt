package com.ehsatn.magik.theme

import android.os.Build
import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.darkColorScheme
import androidx.compose.material3.dynamicDarkColorScheme
import androidx.compose.material3.dynamicLightColorScheme
import androidx.compose.material3.lightColorScheme
import androidx.compose.runtime.Composable
import androidx.compose.ui.platform.LocalContext

private val DarkColorScheme = darkColorScheme(
    primary = MagikOrange,
    secondary = MagikOrange,
    tertiary = MagikOrange,
    background = MagikDarkBackground,
    surface = MagikDarkSurface,
    onPrimary = MagikDarkBackground,
    onSecondary = MagikDarkBackground,
    onTertiary = MagikDarkBackground,
    onBackground = MagikTextPrimary,
    onSurface = MagikTextPrimary,
    error = MagikError,
    onError = MagikDarkBackground
)

@Composable
fun MagikTheme(
  // We force dark theme for MAGIK aesthetic
  darkTheme: Boolean = true,
  dynamicColor: Boolean = false,
  content: @Composable () -> Unit,
) {
  MaterialTheme(colorScheme = DarkColorScheme, typography = Typography, content = content)
}
