package dev.horobets.neobanknative.features.cards

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.AcUnit
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.CheckCircle
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
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
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import dev.horobets.neobanknative.core.design.BrandBackground
import dev.horobets.neobanknative.core.design.PrimaryButton
import dev.horobets.neobanknative.core.design.StatusPill
import dev.horobets.neobanknative.core.design.surfaceCard
import dev.horobets.neobanknative.core.formatting.LedgerAmount
import dev.horobets.neobanknative.core.networking.ApiError
import kotlinx.coroutines.launch

/**
 * Rendered with [initialCard] until [CardsController]'s live list resolves
 * (or if this card somehow drops out of it) — same fallback the wallet
 * screen uses when navigated to with an in-memory copy from the list.
 */
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun CardDetailScreen(cardId: String, initialCard: BankCard, onBack: () -> Unit) {
    val cardsController = LocalCardsController.current
    val state by cardsController.state.collectAsStateWithLifecycle()
    var isSubmitting by remember { mutableStateOf(false) }
    var errorMessage by remember { mutableStateOf<String?>(null) }
    val scope = rememberCoroutineScope()

    val card = (state as? CardsLoadState.Loaded)?.cards?.firstOrNull { it.id == cardId } ?: initialCard

    fun toggleFreeze() {
        if (isSubmitting) return
        errorMessage = null
        isSubmitting = true
        scope.launch {
            try {
                cardsController.toggleFreeze(card)
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
                title = { Text("•••• ${card.lastFour}") },
                navigationIcon = {
                    IconButton(onClick = onBack) { Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Back") }
                },
                colors = TopAppBarDefaults.topAppBarColors(containerColor = Color.Transparent),
            )
        },
    ) { padding ->
        Box(modifier = Modifier.fillMaxSize().padding(padding)) {
            BrandBackground()

            Column(modifier = Modifier.fillMaxSize().padding(20.dp), verticalArrangement = Arrangement.spacedBy(20.dp)) {
                Column(
                    modifier = Modifier.fillMaxWidth().surfaceCard().padding(20.dp),
                    verticalArrangement = Arrangement.spacedBy(10.dp),
                ) {
                    Text("•••• •••• •••• ${card.lastFour}", style = MaterialTheme.typography.titleLarge, fontWeight = FontWeight.Bold)
                    Text(
                        "Expires ${card.expiry}",
                        style = MaterialTheme.typography.bodyMedium,
                        color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f),
                    )
                    StatusPill(
                        text = card.status.replaceFirstChar { it.uppercase() },
                        icon = if (card.isFrozen) Icons.Filled.AcUnit else Icons.Filled.CheckCircle,
                        tint = if (card.isFrozen) MaterialTheme.colorScheme.primary else Color(0xFF22C55E),
                    )
                    card.dailyLimit?.let {
                        Text(
                            "Daily limit: \$${LedgerAmount.formatted(it)}",
                            style = MaterialTheme.typography.bodySmall,
                            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f),
                        )
                    }
                    Text(
                        if (card.onlineOnly) "Online purchases only" else "Online + in-person",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f),
                    )
                }

                errorMessage?.let {
                    Text(it, color = MaterialTheme.colorScheme.error, style = MaterialTheme.typography.bodySmall)
                }

                PrimaryButton(onClick = ::toggleFreeze, modifier = Modifier.fillMaxWidth(), isLoading = isSubmitting) {
                    Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                        Icon(Icons.Filled.AcUnit, contentDescription = null, tint = Color.White)
                        Text(if (card.isFrozen) "Unfreeze card" else "Freeze card", fontWeight = FontWeight.SemiBold)
                    }
                }
            }
        }
    }
}
