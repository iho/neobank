package dev.horobets.neobanknative.features.auth

import dev.horobets.neobanknative.core.networking.ApiClient
import dev.horobets.neobanknative.core.networking.EmptyResponse
import dev.horobets.neobanknative.core.networking.HttpMethod
import dev.horobets.neobanknative.core.networking.jsonBodyOf

/**
 * Wraps the two auth-adjacent clients: [authClient] for the unauthenticated
 * `/v1/auth/login` and `/v1/auth/register` endpoints, [client] for calls
 * that need (and may transparently refresh) a bearer token.
 */
class AuthRepository(
    private val authClient: ApiClient,
    private val client: ApiClient,
) {
    suspend fun login(email: String, password: String): AuthTokens = authClient.send(
        "/v1/auth/login",
        method = HttpMethod.POST,
        body = jsonBodyOf("email" to email, "password" to password),
    )

    suspend fun register(email: String, password: String, phone: String?, inviteCode: String?): AuthTokens = authClient.send(
        "/v1/auth/register",
        method = HttpMethod.POST,
        body = jsonBodyOf(
            "email" to email,
            "password" to password,
            "phone" to phone?.ifEmpty { null },
            "invite_code" to inviteCode?.ifEmpty { null },
        ),
    )

    suspend fun profile(): Profile = client.send("/v1/me")

    suspend fun changePassword(current: String, new: String) {
        client.send<EmptyResponse>(
            "/v1/auth/change-password",
            method = HttpMethod.POST,
            body = jsonBodyOf("current_password" to current, "new_password" to new),
        )
    }
}
