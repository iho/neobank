package dev.horobets.neobanknative.features.cards

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.AcUnit
import androidx.compose.material.icons.filled.ChevronRight
import androidx.compose.material.icons.filled.CreditCard
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.material3.TopAppBarDefaults
import androidx.compose.material3.pulltorefresh.PullToRefreshBox
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import dev.horobets.neobanknative.core.auth.LocalSessionStore
import dev.horobets.neobanknative.core.design.BrandBackground
import dev.horobets.neobanknative.core.design.GlowIcon
import dev.horobets.neobanknative.core.design.PrimaryButton
import dev.horobets.neobanknative.core.design.surfaceCard
import dev.horobets.neobanknative.core.networking.ApiError
import kotlinx.coroutines.launch

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun CardsListScreen(onIssueCard: () -> Unit, onCardSelected: (BankCard) -> Unit) {
    val sessionStore = LocalSessionStore.current
    val cardsController = LocalCardsController.current
    val generation by sessionStore.generation.collectAsStateWithLifecycle()
    val state by cardsController.state.collectAsStateWithLifecycle()
    val scope = rememberCoroutineScope()

    LaunchedEffect(generation) { cardsController.load() }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Cards") },
                actions = {
                    IconButton(onClick = onIssueCard) {
                        Icon(Icons.Filled.Add, contentDescription = "Issue a card")
                    }
                },
                colors = TopAppBarDefaults.topAppBarColors(containerColor = Color.Transparent),
            )
        },
    ) { padding ->
        Box(modifier = Modifier.fillMaxSize().padding(padding)) {
            BrandBackground()

            when (val current = state) {
                is CardsLoadState.Loading -> CircularProgressIndicator(modifier = Modifier.align(Alignment.Center))
                is CardsLoadState.Failed -> CardsErrorView(
                    error = current.error,
                    onRetry = { scope.launch { cardsController.load() } },
                )
                is CardsLoadState.Loaded -> if (current.cards.isEmpty()) {
                    EmptyCardsView(onIssueCard, modifier = Modifier.align(Alignment.Center))
                } else {
                    CardsList(current.cards, cardsController, onCardSelected)
                }
            }
        }
    }
}

@Composable
private fun EmptyCardsView(onIssueCard: () -> Unit, modifier: Modifier = Modifier) {
    Column(
        modifier = modifier.padding(24.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.spacedBy(16.dp),
    ) {
        GlowIcon(icon = Icons.Filled.CreditCard, diameter = 72.dp, iconSize = 32.dp)
        Text("No cards yet", style = MaterialTheme.typography.titleLarge, fontWeight = FontWeight.Bold)
        PrimaryButton(onClick = onIssueCard, modifier = Modifier.padding(horizontal = 40.dp)) {
            Text("Issue a card", fontWeight = FontWeight.SemiBold)
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun CardsList(cards: List<BankCard>, cardsController: CardsController, onCardSelected: (BankCard) -> Unit) {
    val scope = rememberCoroutineScope()
    var isRefreshing by remember { mutableStateOf(false) }

    PullToRefreshBox(
        isRefreshing = isRefreshing,
        onRefresh = {
            scope.launch {
                isRefreshing = true
                cardsController.load()
                isRefreshing = false
            }
        },
    ) {
        LazyColumn(
            modifier = Modifier.fillMaxSize(),
            contentPadding = PaddingValues(20.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            items(cards, key = { it.id }) { card ->
                CardRow(card, onClick = { onCardSelected(card) })
            }
        }
    }
}

@Composable
private fun CardRow(card: BankCard, onClick: () -> Unit) {
    val tint = if (card.isFrozen) MaterialTheme.colorScheme.primary else Color(0xFF22C55E)

    Row(
        modifier = Modifier
            .fillMaxWidth()
            .surfaceCard(cornerRadius = 14.dp)
            .clickable(onClick = onClick)
            .padding(12.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(14.dp),
    ) {
        Box(
            modifier = Modifier.size(40.dp).clip(CircleShape).background(tint.copy(alpha = 0.15f)),
            contentAlignment = Alignment.Center,
        ) {
            Icon(if (card.isFrozen) Icons.Filled.AcUnit else Icons.Filled.CreditCard, contentDescription = null, tint = tint)
        }

        Column(modifier = Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
            Text("•••• ${card.lastFour}", fontWeight = FontWeight.Medium)
            Text(
                "${card.status.replaceFirstChar { it.uppercase() }} · exp ${card.expiry}",
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f),
                maxLines = 1,
            )
        }

        Icon(Icons.Filled.ChevronRight, contentDescription = null, tint = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f))
    }
}

@Composable
private fun CardsErrorView(error: ApiError, onRetry: () -> Unit) {
    Column(
        modifier = Modifier.fillMaxSize().padding(24.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center,
    ) {
        Text(error.message ?: "Something went wrong.", textAlign = TextAlign.Center, modifier = Modifier.fillMaxWidth())
        OutlinedButton(onClick = onRetry, modifier = Modifier.padding(top = 12.dp)) { Text("Retry") }
    }
}
