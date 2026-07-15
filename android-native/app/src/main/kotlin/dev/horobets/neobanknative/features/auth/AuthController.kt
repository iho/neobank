package dev.horobets.neobanknative.features.auth

import androidx.compose.runtime.staticCompositionLocalOf
import dev.horobets.neobanknative.core.auth.SessionStatus
import dev.horobets.neobanknative.core.auth.SessionStore
import dev.horobets.neobanknative.core.storage.TokenStorage

/**
 * Drives the auth lifecycle: reading the persisted session on launch, and
 * performing login/register/logout. Screens own their own submitting/error
 * state; this just owns the actions and the resulting session transition.
 */
class AuthController(
    private val repository: AuthRepository,
    private val tokenStorage: TokenStorage,
    private val sessionStore: SessionStore,
) {
    fun bootstrap() {
        val hasToken = tokenStorage.readAccessToken() != null
        sessionStore.setStatus(if (hasToken) SessionStatus.AUTHENTICATED else SessionStatus.UNAUTHENTICATED)
    }

    suspend fun login(email: String, password: String) {
        val tokens = repository.login(email, password)
        tokenStorage.saveTokens(tokens.accessToken, tokens.refreshToken)
        sessionStore.setStatus(SessionStatus.AUTHENTICATED)
    }

    suspend fun register(email: String, password: String, phone: String?, inviteCode: String?) {
        val tokens = repository.register(email, password, phone, inviteCode)
        tokenStorage.saveTokens(tokens.accessToken, tokens.refreshToken)
        sessionStore.setStatus(SessionStatus.AUTHENTICATED)
    }

    fun logout() {
        tokenStorage.clear()
        sessionStore.setStatus(SessionStatus.UNAUTHENTICATED)
    }
}

/** Android's equivalent of SwiftUI's `.environment(_:)` / `@Environment(Type.self)`. */
val LocalAuthController = staticCompositionLocalOf<AuthController> {
    error("No AuthController provided — wrap the composition in AppEnvironment's providers")
}
