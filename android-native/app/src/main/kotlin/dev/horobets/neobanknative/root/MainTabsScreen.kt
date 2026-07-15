package dev.horobets.neobanknative.root

import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.AccountBalanceWallet
import androidx.compose.material.icons.filled.CreditCard
import androidx.compose.material.icons.filled.Notifications
import androidx.compose.material3.Icon
import androidx.compose.material3.NavigationBar
import androidx.compose.material3.NavigationBarItem
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.foundation.layout.padding
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.navigation.NavGraph.Companion.findStartDestination
import androidx.navigation.NavType
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.currentBackStackEntryAsState
import androidx.navigation.compose.rememberNavController
import androidx.navigation.navArgument
import dev.horobets.neobanknative.features.cards.CardDetailScreen
import dev.horobets.neobanknative.features.cards.CardsListScreen
import dev.horobets.neobanknative.features.cards.CardsLoadState
import dev.horobets.neobanknative.features.cards.IssueCardScreen
import dev.horobets.neobanknative.features.cards.LocalCardsController
import dev.horobets.neobanknative.features.notifications.NotificationsListScreen
import dev.horobets.neobanknative.features.transfers.TransferFlowScreen
import dev.horobets.neobanknative.features.wallet.WalletScreen

private const val ROUTE_WALLET = "wallet"
private const val ROUTE_CARDS = "cards"
private const val ROUTE_ALERTS = "alerts"
private const val ROUTE_ISSUE_CARD = "issueCard"
private const val ROUTE_SEND_MONEY = "sendMoney"
private const val ROUTE_CARD_DETAIL = "cardDetail/{cardId}"

private data class TabDestination(val route: String, val label: String, val icon: ImageVector)

private val tabs = listOf(
    TabDestination(ROUTE_WALLET, "Wallet", Icons.Filled.AccountBalanceWallet),
    TabDestination(ROUTE_CARDS, "Cards", Icons.Filled.CreditCard),
    TabDestination(ROUTE_ALERTS, "Alerts", Icons.Filled.Notifications),
)

/**
 * Each tab renders its own `Scaffold`/top bar (mirroring ios-native's
 * `TabView` of independent `NavigationStack`s); card detail, issue-card,
 * and send-money are pushed as full-screen routes on the same [NavHost]
 * rather than presented as sheets, the idiomatic Android equivalent of
 * ios-native's `.sheet(isPresented:)` modals.
 */
@Composable
fun MainTabsScreen() {
    val navController = rememberNavController()
    val backStackEntry by navController.currentBackStackEntryAsState()
    val currentRoute = backStackEntry?.destination?.route

    Scaffold(
        bottomBar = {
            if (currentRoute == null || tabs.any { it.route == currentRoute }) {
                NavigationBar {
                    tabs.forEach { tab ->
                        NavigationBarItem(
                            selected = currentRoute == tab.route,
                            onClick = {
                                navController.navigate(tab.route) {
                                    popUpTo(navController.graph.findStartDestination().id) { saveState = true }
                                    launchSingleTop = true
                                    restoreState = true
                                }
                            },
                            icon = { Icon(tab.icon, contentDescription = tab.label) },
                            label = { Text(tab.label) },
                        )
                    }
                }
            }
        },
    ) { padding ->
        NavHost(
            navController = navController,
            startDestination = ROUTE_WALLET,
            modifier = Modifier.padding(padding),
        ) {
            composable(ROUTE_WALLET) {
                WalletScreen(onSendMoney = { navController.navigate(ROUTE_SEND_MONEY) })
            }
            composable(ROUTE_CARDS) {
                CardsListScreen(
                    onIssueCard = { navController.navigate(ROUTE_ISSUE_CARD) },
                    onCardSelected = { card -> navController.navigate("cardDetail/${card.id}") },
                )
            }
            composable(ROUTE_ALERTS) {
                NotificationsListScreen()
            }
            composable(
                route = ROUTE_CARD_DETAIL,
                arguments = listOf(navArgument("cardId") { type = NavType.StringType }),
            ) { entry ->
                val cardId = entry.arguments?.getString("cardId").orEmpty()
                val cardsController = LocalCardsController.current
                val cardsState by cardsController.state.collectAsStateWithLifecycle()
                val initialCard = (cardsState as? CardsLoadState.Loaded)?.cards?.firstOrNull { it.id == cardId }
                if (initialCard != null) {
                    CardDetailScreen(cardId = cardId, initialCard = initialCard, onBack = { navController.popBackStack() })
                }
            }
            composable(ROUTE_ISSUE_CARD) {
                IssueCardScreen(onDismiss = { navController.popBackStack() })
            }
            composable(ROUTE_SEND_MONEY) {
                TransferFlowScreen(onDismiss = { navController.popBackStack() })
            }
        }
    }
}
