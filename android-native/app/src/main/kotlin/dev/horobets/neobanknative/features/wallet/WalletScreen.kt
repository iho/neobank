package dev.horobets.neobanknative.features.wallet

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.Send
import androidx.compose.material.icons.filled.ArrowDownward
import androidx.compose.material.icons.filled.ArrowUpward
import androidx.compose.material.icons.automirrored.filled.Logout
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.DropdownMenu
import androidx.compose.material3.DropdownMenuItem
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.HorizontalDivider
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
import androidx.compose.runtime.snapshotFlow
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import dev.horobets.neobanknative.core.auth.LocalSessionStore
import dev.horobets.neobanknative.core.design.AppAppearance
import dev.horobets.neobanknative.core.design.BrandBackground
import dev.horobets.neobanknative.core.design.PrimaryButton
import dev.horobets.neobanknative.core.design.surfaceCard
import dev.horobets.neobanknative.core.formatting.DateFormatting
import dev.horobets.neobanknative.core.design.LocalAppAppearanceStore
import dev.horobets.neobanknative.core.formatting.LedgerAmount
import dev.horobets.neobanknative.core.networking.ApiError
import dev.horobets.neobanknative.features.auth.LocalAuthController
import kotlinx.coroutines.launch

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun WalletScreen(onSendMoney: () -> Unit) {
    val sessionStore = LocalSessionStore.current
    val authController = LocalAuthController.current
    val walletController = LocalWalletController.current
    val appAppearanceStore = LocalAppAppearanceStore.current

    val generation by sessionStore.generation.collectAsStateWithLifecycle()
    val state by walletController.state.collectAsStateWithLifecycle()
    val appearance by appAppearanceStore.appearance.collectAsStateWithLifecycle()
    var menuExpanded by remember { mutableStateOf(false) }
    val scope = rememberCoroutineScope()

    LaunchedEffect(generation) { walletController.load() }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Wallet") },
                actions = {
                    Box {
                        IconButton(onClick = { menuExpanded = true }) {
                            Icon(appearance.icon, contentDescription = "Settings")
                        }
                        DropdownMenu(expanded = menuExpanded, onDismissRequest = { menuExpanded = false }) {
                            AppAppearance.entries.forEach { option ->
                                DropdownMenuItem(
                                    text = { Text(option.label) },
                                    leadingIcon = { Icon(option.icon, contentDescription = null) },
                                    onClick = {
                                        appAppearanceStore.setAppearance(option)
                                        menuExpanded = false
                                    },
                                )
                            }
                            HorizontalDivider()
                            DropdownMenuItem(
                                text = { Text("Log out", color = MaterialTheme.colorScheme.error) },
                                leadingIcon = {
                                    Icon(Icons.AutoMirrored.Filled.Logout, contentDescription = null, tint = MaterialTheme.colorScheme.error)
                                },
                                onClick = {
                                    menuExpanded = false
                                    authController.logout()
                                },
                            )
                        }
                    }
                },
                colors = TopAppBarDefaults.topAppBarColors(containerColor = Color.Transparent),
            )
        },
        bottomBar = {
            if (state is WalletLoadState.Loaded) {
                Box(modifier = Modifier.fillMaxWidth().padding(horizontal = 20.dp, vertical = 8.dp)) {
                    PrimaryButton(onClick = onSendMoney, modifier = Modifier.fillMaxWidth()) {
                        Row(verticalAlignment = Alignment.CenterVertically) {
                            Icon(Icons.AutoMirrored.Filled.Send, contentDescription = null, tint = Color.White)
                            Spacer(Modifier.width(8.dp))
                            Text("Send", fontWeight = FontWeight.SemiBold)
                        }
                    }
                }
            }
        },
    ) { padding ->
        Box(modifier = Modifier.fillMaxSize().padding(padding)) {
            BrandBackground()

            when (val current = state) {
                is WalletLoadState.Loading -> CircularProgressIndicator(modifier = Modifier.align(Alignment.Center))
                is WalletLoadState.Failed -> WalletErrorView(
                    error = current.error,
                    onRetry = { scope.launch { walletController.load() } },
                )
                is WalletLoadState.Loaded -> WalletContent(current.snapshot, walletController)
            }
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun WalletContent(snapshot: WalletSnapshot, walletController: WalletController) {
    val scope = rememberCoroutineScope()
    var isRefreshing by remember { mutableStateOf(false) }
    val listState = rememberLazyListState()

    LaunchedEffect(listState, snapshot.transactions.size) {
        snapshotFlow { listState.layoutInfo.visibleItemsInfo.lastOrNull()?.index }
            .collect { lastVisible ->
                if (lastVisible != null && snapshot.transactions.isNotEmpty() && lastVisible >= snapshot.transactions.lastIndex) {
                    walletController.loadMore()
                }
            }
    }

    PullToRefreshBox(
        isRefreshing = isRefreshing,
        onRefresh = {
            scope.launch {
                isRefreshing = true
                walletController.load()
                isRefreshing = false
            }
        },
    ) {
        LazyColumn(
            state = listState,
            modifier = Modifier.fillMaxSize(),
            contentPadding = PaddingValues(20.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            item { BalanceCard(snapshot.balance) }

            item {
                Text("Transactions", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.SemiBold)
            }

            if (snapshot.transactions.isEmpty()) {
                item {
                    Box(modifier = Modifier.fillMaxWidth().padding(vertical = 32.dp), contentAlignment = Alignment.Center) {
                        Text(
                            "No transactions yet",
                            style = MaterialTheme.typography.bodyMedium,
                            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f),
                        )
                    }
                }
            } else {
                itemsIndexed(snapshot.transactions, key = { _, tx -> tx.id }) { _, transaction ->
                    TransactionRow(transaction)
                }
            }

            if (snapshot.isLoadingMore) {
                item {
                    Box(modifier = Modifier.fillMaxWidth().padding(vertical = 12.dp), contentAlignment = Alignment.Center) {
                        CircularProgressIndicator()
                    }
                }
            }
        }
    }
}

@Composable
private fun BalanceCard(balance: WalletBalance) {
    Column(
        modifier = Modifier.fillMaxWidth().surfaceCard().padding(20.dp),
        verticalArrangement = Arrangement.spacedBy(8.dp),
    ) {
        Text(
            "Available balance",
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f),
        )
        Text(
            "${LedgerAmount.formatted(balance.availableBalance)} ${balance.currency}",
            style = MaterialTheme.typography.headlineMedium,
            fontWeight = FontWeight.Bold,
        )
        balance.encumberedBalance?.let {
            Text(
                "Held: ${LedgerAmount.formatted(it)} ${balance.currency}",
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f),
            )
        }
    }
}

@Composable
private fun TransactionRow(transaction: WalletTransaction) {
    val tint = if (transaction.isCredit) Color(0xFF22C55E) else MaterialTheme.colorScheme.onSurface

    Row(
        modifier = Modifier.fillMaxWidth().surfaceCard(cornerRadius = 14.dp).padding(12.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(14.dp),
    ) {
        Box(
            modifier = Modifier
                .size(40.dp)
                .clip(CircleShape)
                .background(tint.copy(alpha = 0.15f)),
            contentAlignment = Alignment.Center,
        ) {
            Icon(
                imageVector = if (transaction.isCredit) Icons.Filled.ArrowDownward else Icons.Filled.ArrowUpward,
                contentDescription = null,
                tint = tint,
            )
        }

        Column(modifier = Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
            Text(transaction.counterparty ?: transaction.type, fontWeight = FontWeight.Medium)
            Text(
                "${transaction.status} · ${DateFormatting.formatted(transaction.createdAt)}",
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f),
            )
        }

        Text(
            "${if (transaction.isCredit) "+" else "-"}${transaction.amount} ${transaction.currency}",
            fontWeight = FontWeight.SemiBold,
            color = tint,
        )
    }
}

@Composable
private fun WalletErrorView(error: ApiError, onRetry: () -> Unit) {
    Column(
        modifier = Modifier.fillMaxSize().padding(24.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center,
    ) {
        Text(error.message ?: "Something went wrong.", textAlign = TextAlign.Center, modifier = Modifier.fillMaxWidth())
        Spacer(Modifier.width(12.dp))
        OutlinedButton(onClick = onRetry) { Text("Retry") }
    }
}

