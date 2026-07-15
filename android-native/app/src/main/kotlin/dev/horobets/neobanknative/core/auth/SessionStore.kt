package dev.horobets.neobanknative.core.auth

import androidx.compose.runtime.staticCompositionLocalOf
import java.util.UUID
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

enum class SessionStatus {
    UNKNOWN,
    AUTHENTICATED,
    UNAUTHENTICATED,
}

/**
 * Session-scoped cache invalidation via a generation token, not an explicit
 * invalidate-list. Rather than every login/logout explicitly resetting each
 * feature ViewModel, [generation] is bumped on every transition; feature
 * screens key their reload `LaunchedEffect(sessionStore.generation)` off
 * that. A ViewModel doesn't need to know it must be reset — it just always
 * reloads fresh data when the session changes, including a
 * logout-then-login-as-someone-else cycle.
 */
class SessionStore {
    private val _status = MutableStateFlow(SessionStatus.UNKNOWN)
    val status: StateFlow<SessionStatus> = _status.asStateFlow()

    private val _generation = MutableStateFlow(UUID.randomUUID())
    val generation: StateFlow<UUID> = _generation.asStateFlow()

    fun setStatus(status: SessionStatus) {
        _status.value = status
        _generation.value = UUID.randomUUID()
    }
}

val LocalSessionStore = staticCompositionLocalOf<SessionStore> {
    error("No SessionStore provided — wrap the composition in AppEnvironment's providers")
}
