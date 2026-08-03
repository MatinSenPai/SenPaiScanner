package com.matinsenpai.senpaiscanner.ui.main

import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.RowScope
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.FilterChip
import androidx.compose.material3.FilterChipDefaults
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Surface
import androidx.compose.material3.Switch
import androidx.compose.material3.SwitchDefaults
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.semantics.Role
import androidx.compose.ui.semantics.heading
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.matinsenpai.senpaiscanner.theme.SignalBorder
import com.matinsenpai.senpaiscanner.theme.SignalCyan
import com.matinsenpai.senpaiscanner.theme.SignalMuted
import com.matinsenpai.senpaiscanner.theme.SignalPanel
import com.matinsenpai.senpaiscanner.theme.SignalPanelRaised
import com.matinsenpai.senpaiscanner.theme.SignalText

private val panelShape = RoundedCornerShape(10.dp)

@Composable
fun DeskPanel(
    eyebrow: String,
    title: String,
    modifier: Modifier = Modifier,
    action: (@Composable () -> Unit)? = null,
    content: @Composable () -> Unit,
) {
    Surface(
        modifier = modifier.fillMaxWidth(),
        color = SignalPanel,
        shape = panelShape,
        border = BorderStroke(1.dp, SignalBorder),
        tonalElevation = 0.dp,
    ) {
        Column(Modifier.padding(14.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Column(Modifier.weight(1f)) {
                    Text(eyebrow, color = SignalCyan, fontSize = 10.sp, fontWeight = FontWeight.Black, letterSpacing = 1.sp)
                    Text(
                        title,
                        color = SignalText,
                        style = MaterialTheme.typography.titleMedium,
                        fontWeight = FontWeight.Bold,
                        modifier = Modifier.semantics { heading() },
                    )
                }
                action?.invoke()
            }
            content()
        }
    }
}

@Composable
fun ToggleSetting(
    title: String,
    supporting: String,
    checked: Boolean,
    onCheckedChange: (Boolean) -> Unit,
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .background(SignalPanelRaised, RoundedCornerShape(8.dp))
            .clickable(role = Role.Switch) { onCheckedChange(!checked) }
            .padding(horizontal = 12.dp, vertical = 9.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Column(Modifier.weight(1f)) {
            Text(title, color = SignalText, fontWeight = FontWeight.SemiBold, fontSize = 14.sp)
            Text(supporting, color = SignalMuted, fontSize = 11.sp, lineHeight = 15.sp)
        }
        Switch(
            checked = checked,
            onCheckedChange = onCheckedChange,
            colors = SwitchDefaults.colors(
                checkedThumbColor = SignalPanel,
                checkedTrackColor = SignalCyan,
                uncheckedThumbColor = SignalMuted,
                uncheckedTrackColor = SignalBorder,
            ),
        )
    }
}

@Composable
fun <T> ChoiceChips(
    choices: List<Pair<String, T>>,
    selected: T,
    onSelected: (T) -> Unit,
) {
    Row(horizontalArrangement = Arrangement.spacedBy(7.dp), modifier = Modifier.fillMaxWidth()) {
        choices.forEach { (label, value) ->
            FilterChip(
                selected = selected == value,
                onClick = { onSelected(value) },
                label = { Text(label, maxLines = 1) },
                colors = FilterChipDefaults.filterChipColors(
                    containerColor = SignalPanelRaised,
                    labelColor = SignalMuted,
                    selectedContainerColor = SignalCyan.copy(alpha = .14f),
                    selectedLabelColor = SignalCyan,
                ),
                border = FilterChipDefaults.filterChipBorder(
                    enabled = true,
                    selected = selected == value,
                    borderColor = SignalBorder,
                    selectedBorderColor = SignalCyan,
                ),
                modifier = Modifier.weight(1f),
            )
        }
    }
}

@Composable
fun RowScope.DeskAction(
    label: String,
    onClick: () -> Unit,
    enabled: Boolean = true,
    accent: Color = SignalCyan,
) {
    OutlinedButton(
        onClick = onClick,
        enabled = enabled,
        modifier = Modifier.weight(1f),
        border = BorderStroke(1.dp, if (enabled) accent else SignalBorder),
        colors = ButtonDefaults.outlinedButtonColors(contentColor = accent, disabledContentColor = SignalMuted),
        shape = RoundedCornerShape(8.dp),
    ) {
        Text(label, fontSize = 12.sp, fontWeight = FontWeight.Bold, maxLines = 1)
    }
}
