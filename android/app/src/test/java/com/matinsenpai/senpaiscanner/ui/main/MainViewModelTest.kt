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
