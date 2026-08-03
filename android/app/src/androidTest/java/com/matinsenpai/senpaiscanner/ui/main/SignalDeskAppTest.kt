package com.matinsenpai.senpaiscanner.ui.main

import androidx.activity.ComponentActivity
import androidx.compose.ui.test.assertIsDisplayed
import androidx.compose.ui.test.junit4.createAndroidComposeRule
import androidx.compose.ui.test.onNodeWithText
import androidx.compose.ui.test.performClick
import com.matinsenpai.senpaiscanner.theme.SenPaiScannerTheme
import org.junit.Rule
import org.junit.Test

class SignalDeskAppTest {
    @get:Rule
    val composeRule = createAndroidComposeRule<ComponentActivity>()

    @Test
    fun resultsTab_exposesLiveCopyAndSpeedActions() {
        val state = ScanUiState(
            phase1Tested = 2,
            phase1Healthy = 1,
            phase1Failed = 1,
            results = listOf(IpResult("1.1.1.1", 443, 42, 0.0, "FRA", true)),
        )
        composeRule.setContent {
            SenPaiScannerTheme {
                SignalDeskApp(
                    uiState = state,
                    onConfigChange = {},
                    onStart = {},
                    onStop = {},
                    onSpeedTest = {},
                    onGenerateExports = {},
                    onDismissError = {},
                )
            }
        }

        composeRule.onNodeWithText("Results").performClick()
        composeRule.onNodeWithText("COPY GREEN").assertIsDisplayed()
        composeRule.onNodeWithText("COPY TOP 20").assertIsDisplayed()
        composeRule.onNodeWithText("SPEED TEST 1 GREEN RESULTS").assertIsDisplayed()
    }

    @Test
    fun exportTab_isSeparateFromResults() {
        composeRule.setContent {
            SenPaiScannerTheme {
                SignalDeskApp(
                    uiState = ScanUiState(),
                    onConfigChange = {},
                    onStart = {},
                    onStop = {},
                    onSpeedTest = {},
                    onGenerateExports = {},
                    onDismissError = {},
                )
            }
        }

        composeRule.onNodeWithText("Export").performClick()
        composeRule.onNodeWithText("RAW ENDPOINTS").assertIsDisplayed()
        composeRule.onNodeWithText("CLIENT CONFIGS").assertIsDisplayed()
    }
}
