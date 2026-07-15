package dev.horobets.neobanknative.features.cards

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Switch
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.material3.TopAppBar
import androidx.compose.material3.TopAppBarDefaults
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.dp
import dev.horobets.neobanknative.core.design.BrandBackground
import dev.horobets.neobanknative.core.design.BrandTextField
import dev.horobets.neobanknative.core.design.PrimaryButton
import dev.horobets.neobanknative.core.design.surfaceCard
import dev.horobets.neobanknative.core.networking.ApiError
import kotlinx.coroutines.launch

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun IssueCardScreen(onDismiss: () -> Unit) {
    val cardsController = LocalCardsController.current
    var cardholderName by remember { mutableStateOf("") }
    var dailyLimit by remember { mutableStateOf("") }
    var onlineOnly by remember { mutableStateOf(false) }
    var isSubmitting by remember { mutableStateOf(false) }
    var errorMessage by remember { mutableStateOf<String?>(null) }
    val scope = rememberCoroutineScope()

    val canSubmit = cardholderName.trim().isNotEmpty()

    fun submit() {
        if (!canSubmit || isSubmitting) return
        errorMessage = null
        isSubmitting = true
        scope.launch {
            try {
                cardsController.issueCard(cardholderName.trim(), dailyLimit.trim(), onlineOnly)
                onDismiss()
            } catch (e: ApiError) {
                errorMessage = e.message
            } catch (e: Exception) {
                errorMessage = "Something went wrong. Please try again."
            } finally {
                isSubmitting = false
            }
        }
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Issue a virtual card") },
                navigationIcon = { TextButton(onClick = onDismiss) { Text("Cancel") } },
                colors = TopAppBarDefaults.topAppBarColors(containerColor = Color.Transparent),
            )
        },
    ) { padding ->
        Box(modifier = Modifier.fillMaxSize().padding(padding)) {
            BrandBackground()

            Column(modifier = Modifier.fillMaxSize().padding(20.dp), verticalArrangement = Arrangement.spacedBy(16.dp)) {
                BrandTextField(
                    value = cardholderName,
                    onValueChange = { cardholderName = it },
                    placeholder = "Cardholder name",
                    modifier = Modifier.fillMaxWidth(),
                )

                BrandTextField(
                    value = dailyLimit,
                    onValueChange = { dailyLimit = it },
                    placeholder = "Daily limit (optional)",
                    modifier = Modifier.fillMaxWidth(),
                    keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Decimal),
                )

                Row(
                    modifier = Modifier.fillMaxWidth().surfaceCard(cornerRadius = 12.dp).padding(14.dp),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    Text("Online purchases only")
                    Switch(checked = onlineOnly, onCheckedChange = { onlineOnly = it })
                }

                errorMessage?.let {
                    Text(it, color = MaterialTheme.colorScheme.error, style = MaterialTheme.typography.bodySmall)
                }

                PrimaryButton(
                    onClick = ::submit,
                    modifier = Modifier.fillMaxWidth(),
                    enabled = canSubmit,
                    isLoading = isSubmitting,
                ) {
                    Text("Issue card", fontWeight = FontWeight.SemiBold)
                }
            }
        }
    }
}
