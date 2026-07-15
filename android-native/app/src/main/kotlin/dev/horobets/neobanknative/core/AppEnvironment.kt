package dev.horobets.neobanknative.core

import android.content.Context
import dev.horobets.neobanknative.core.auth.SessionStatus
import dev.horobets.neobanknative.core.auth.SessionStore
import dev.horobets.neobanknative.core.config.AppConfig
import dev.horobets.neobanknative.core.design.AppAppearanceStore
import dev.horobets.neobanknative.core.networking.ApiClient
import dev.horobets.neobanknative.core.networking.HttpMethod
import dev.horobets.neobanknative.core.networking.jsonBodyOf
import dev.horobets.neobanknative.core.storage.TokenStorage
import dev.horobets.neobanknative.features.auth.AuthController
import dev.horobets.neobanknative.features.auth.AuthRepository
import dev.horobets.neobanknative.features.cards.CardRepository
import dev.horobets.neobanknative.features.cards.CardsController
import dev.horobets.neobanknative.features.kyc.KycController
import dev.horobets.neobanknative.features.kyc.KycRepository
import dev.horobets.neobanknative.features.notifications.NotificationRepository
import dev.horobets.neobanknative.features.notifications.NotificationsController
import dev.horobets.neobanknative.features.transfers.TransferRepository
import dev.horobets.neobanknative.features.transfers.TransferSubmitController
import dev.horobets.neobanknative.features.wallet.WalletController
import dev.horobets.neobanknative.features.wallet.WalletRepository
import kotlinx.serialization.Serializable

/**
 * Composition root: wires the unauthenticated auth client, the token-bearing
 * client (with refresh-on-401), and the controllers built on top of them —
 * Android's equivalent of ios-native's `AppEnvironment`, minus the SwiftUI
 * `@main` scene; see [dev.horobets.neobanknative.NeobankApplication] for
 * where this gets instantiated once for the process lifetime.
 */
class AppEnvironment(context: Context) {
    @Serializable
    private data class RefreshResponse(val accessToken: String, val refreshToken: String)

    val sessionStore = SessionStore()
    val appAppearanceStore = AppAppearanceStore(context)

    val authController: AuthController
    val kycController: KycController
    val walletController: WalletController
    val cardsController: CardsController
    val notificationsController: NotificationsController
    val transferSubmitController: TransferSubmitController

    init {
        val tokenStorage = TokenStorage(context)
        val authClient = ApiClient(baseUrl = AppConfig.apiBaseUrl)

        val apiClient = ApiClient(
            baseUrl = AppConfig.apiBaseUrl,
            auth = ApiClient.AuthContext(
                tokenProvider = { tokenStorage.readAccessToken() },
                refresh = refresh@{
                    val refreshToken = tokenStorage.readRefreshToken() ?: return@refresh null
                    val response = try {
                        authClient.send<RefreshResponse>(
                            "/v1/auth/refresh",
                            method = HttpMethod.POST,
                            body = jsonBodyOf("refresh_token" to refreshToken),
                        )
                    } catch (e: Exception) {
                        return@refresh null
                    }
                    tokenStorage.saveTokens(response.accessToken, response.refreshToken)
                    response.accessToken
                },
                onSessionExpired = {
                    tokenStorage.clear()
                    sessionStore.setStatus(SessionStatus.UNAUTHENTICATED)
                },
            ),
        )

        val authRepository = AuthRepository(authClient, apiClient)
        val kycRepository = KycRepository(apiClient)
        val walletRepository = WalletRepository(apiClient)
        val cardRepository = CardRepository(apiClient)
        val notificationRepository = NotificationRepository(apiClient)
        val transferRepository = TransferRepository(apiClient)

        authController = AuthController(authRepository, tokenStorage, sessionStore)
        kycController = KycController(kycRepository)
        walletController = WalletController(walletRepository)
        cardsController = CardsController(cardRepository)
        notificationsController = NotificationsController(notificationRepository)
        transferSubmitController = TransferSubmitController(transferRepository)
    }
}
