package com.matinsenpai.senpaiscanner.ui.main

import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.Json
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Test

class MainViewModelTest {
    @Test
    fun neighborScan_isOptInAndSerialized() {
        val config = ScanConfig()

        assertFalse(config.neighborScan)
        assertEquals(false, Json.decodeFromString<ScanConfig>(Json.encodeToString(config)).neighborScan)
    }

    // The engine reads the JSON the app sends. Defaults must be sent too: a missing "requireWebSocket" or
    // "antiDpiEnabled" would reach Go as false and silently switch the option off.
    @Test
    fun scanConfig_sendsDefaultsAndNewOptions() {
        val sender = Json { encodeDefaults = true }
        val sent = sender.encodeToString(ScanConfig(gentle = true, targets = "1.1.1.1", phase2Only = true))
        val back = Json.decodeFromString<ScanConfig>(sent)

        assertEquals(true, back.requireWebSocket)
        assertEquals(true, back.antiDpiEnabled)
        assertEquals(true, back.gentle)
        assertEquals("1.1.1.1", back.targets)
        assertEquals(true, back.phase2Only)
        assertFalse(back.resume)
        assertEquals(true, sent.contains("\"antiDpiEnabled\":true"))
    }

    @Test
    fun resumeInfo_decodesTheEnginePayload() {
        val info = Json { ignoreUnknownKeys = true }.decodeFromString<ResumeInfo>(
            """{"saved":"2026-10-04T01:00:00Z","tested":120,"total":5000,"healthy":7,"phase":1}""",
        )

        assertEquals(120, info.tested)
        assertEquals(5000, info.total)
        assertEquals(7, info.healthy)
    }

    @Test
    fun healthyEndpoints_areUniqueSortedAndLimited() {
        val input = listOf(
            IpResult("1.1.1.1", 443, 80, 0.0, "FRA", true),
            IpResult("2.2.2.2", 443, 20, 0.0, "AMS", true),
            IpResult("1.1.1.1", 443, 10, 0.0, "FRA", true),
            IpResult("3.3.3.3", 443, 5, 0.0, "LHR", true, isPhase2 = true),
        )

        val result = healthyEndpoints(input, limit = 2)

        assertEquals(listOf("1.1.1.1", "2.2.2.2"), result.map { it.ip })
        assertEquals(listOf(10, 20), result.map { it.latencyMs })
    }
}
