package dev.horobets.neobanknative.features.notifications

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
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
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Notifications
import androidx.compose.material.icons.filled.NotificationsActive
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
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
import dev.horobets.neobanknative.core.design.surfaceCard
import dev.horobets.neobanknative.core.formatting.DateFormatting
import dev.horobets.neobanknative.core.networking.ApiError
import kotlinx.coroutines.delay
import kotlinx.coroutines.isActive
import kotlinx.coroutines.launch

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun NotificationsListScreen() {
    val sessionStore = LocalSessionStore.current
    val notificationsController = LocalNotificationsController.current
    val generation by sessionStore.generation.collectAsStateWithLifecycle()
    val state by notificationsController.state.collectAsStateWithLifecycle()
    val scope = rememberCoroutineScope()

    LaunchedEffect(generation) {
        notificationsController.load()
        while (isActive) {
            delay(30_000)
            notificationsController.refresh()
        }
    }

    val unreadCount = (state as? NotificationsLoadState.Loaded)?.page?.unreadCount ?: 0

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Alerts") },
                actions = {
                    TextButton(
                        onClick = { scope.launch { runCatching { notificationsController.markAllRead() } } },
                        enabled = unreadCount > 0,
                    ) { Text("Mark all read") }
                },
                colors = TopAppBarDefaults.topAppBarColors(containerColor = Color.Transparent),
            )
        },
    ) { padding ->
        Box(modifier = Modifier.fillMaxSize().padding(padding)) {
            BrandBackground()

            when (val current = state) {
                is NotificationsLoadState.Loading -> CircularProgressIndicator(modifier = Modifier.align(Alignment.Center))
                is NotificationsLoadState.Failed -> NotificationsErrorView(
                    error = current.error,
                    onRetry = { scope.launch { notificationsController.load() } },
                )
                is NotificationsLoadState.Loaded -> if (current.page.notifications.isEmpty()) {
                    EmptyNotificationsView(modifier = Modifier.align(Alignment.Center))
                } else {
                    NotificationsList(current.page.notifications, notificationsController)
                }
            }
        }
    }
}

@Composable
private fun EmptyNotificationsView(modifier: Modifier = Modifier) {
    Column(modifier = modifier, horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(16.dp)) {
        GlowIcon(icon = Icons.Filled.Notifications, diameter = 72.dp, iconSize = 32.dp)
        Text("No notifications yet", style = MaterialTheme.typography.titleLarge, fontWeight = FontWeight.Bold)
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun NotificationsList(notifications: List<AppNotification>, notificationsController: NotificationsController) {
    val scope = rememberCoroutineScope()
    var isRefreshing by remember { mutableStateOf(false) }

    PullToRefreshBox(
        isRefreshing = isRefreshing,
        onRefresh = {
            scope.launch {
                isRefreshing = true
                notificationsController.load()
                isRefreshing = false
            }
        },
    ) {
        LazyColumn(
            modifier = Modifier.fillMaxSize(),
            contentPadding = PaddingValues(20.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            items(notifications, key = { it.id }) { notification ->
                NotificationRow(notification) {
                    scope.launch { runCatching { notificationsController.markRead(notification.id) } }
                }
            }
        }
    }
}

@Composable
private fun NotificationRow(notification: AppNotification, onTap: () -> Unit) {
    val tint = if (notification.read) MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f) else MaterialTheme.colorScheme.primary

    Row(
        modifier = Modifier
            .fillMaxWidth()
            .surfaceCard(cornerRadius = 14.dp)
            .clickable(enabled = !notification.read, onClick = onTap)
            .padding(12.dp),
        verticalAlignment = Alignment.Top,
        horizontalArrangement = Arrangement.spacedBy(14.dp),
    ) {
        Box(
            modifier = Modifier.size(40.dp).clip(CircleShape).background(tint.copy(alpha = 0.15f)),
            contentAlignment = Alignment.Center,
        ) {
            Icon(
                if (notification.read) Icons.Filled.Notifications else Icons.Filled.NotificationsActive,
                contentDescription = null,
                tint = tint,
            )
        }

        Column(modifier = Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(4.dp)) {
            Text(notification.title, fontWeight = if (notification.read) FontWeight.Normal else FontWeight.Bold)
            Text(
                notification.body,
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f),
                maxLines = 2,
            )
            Text(
                DateFormatting.formatted(notification.createdAt),
                style = MaterialTheme.typography.labelSmall,
                color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f),
            )
        }
    }
}

@Composable
private fun NotificationsErrorView(error: ApiError, onRetry: () -> Unit) {
    Column(
        modifier = Modifier.fillMaxSize().padding(24.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center,
    ) {
        Text(error.message ?: "Something went wrong.", textAlign = TextAlign.Center, modifier = Modifier.fillMaxWidth())
        Spacer(Modifier.padding(6.dp))
        OutlinedButton(onClick = onRetry) { Text("Retry") }
    }
}
