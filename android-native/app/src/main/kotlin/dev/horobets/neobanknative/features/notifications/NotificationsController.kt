package dev.horobets.neobanknative.features.notifications

import androidx.compose.runtime.staticCompositionLocalOf
import dev.horobets.neobanknative.core.networking.ApiError
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

sealed interface NotificationsLoadState {
    data object Loading : NotificationsLoadState
    data class Loaded(val page: NotificationPage) : NotificationsLoadState
    data class Failed(val error: ApiError) : NotificationsLoadState
}

class NotificationsController(private val repository: NotificationRepository) {
    private val _state = MutableStateFlow<NotificationsLoadState>(NotificationsLoadState.Loading)
    val state: StateFlow<NotificationsLoadState> = _state.asStateFlow()

    suspend fun load() {
        _state.value = NotificationsLoadState.Loading
        _state.value = try {
            NotificationsLoadState.Loaded(repository.list())
        } catch (e: ApiError) {
            NotificationsLoadState.Failed(e)
        } catch (e: Exception) {
            NotificationsLoadState.Failed(ApiError.decoding())
        }
    }

    /** Used by the background poll loop: keeps the last good page on screen instead of flashing a spinner or error every 30 seconds on a blip. */
    suspend fun refresh() {
        val page = try {
            repository.list()
        } catch (e: Exception) {
            return
        }
        _state.value = NotificationsLoadState.Loaded(page)
    }

    suspend fun markRead(id: String) {
        val loaded = _state.value as? NotificationsLoadState.Loaded ?: return
        val updated = repository.markRead(id)
        val notifications = loaded.page.notifications.map { if (it.id == id) updated else it }
        val newUnread = (loaded.page.unreadCount - if (updated.read) 1 else 0).coerceAtLeast(0)
        _state.value = NotificationsLoadState.Loaded(loaded.page.copy(notifications = notifications, unreadCount = newUnread))
    }

    suspend fun markAllRead() {
        val loaded = _state.value as? NotificationsLoadState.Loaded ?: return
        repository.markAllRead()
        val notifications = loaded.page.notifications.map { it.copy(read = true) }
        _state.value = NotificationsLoadState.Loaded(loaded.page.copy(notifications = notifications, unreadCount = 0))
    }
}

val LocalNotificationsController = staticCompositionLocalOf<NotificationsController> {
    error("No NotificationsController provided — wrap the composition in AppEnvironment's providers")
}
