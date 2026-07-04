import Foundation

/// Composition root: wires the unauthenticated auth client, the token-bearing
/// client (with refresh-on-401), and the objects built on top of them.
@MainActor
struct AppEnvironment {
    let sessionStore: SessionStore
    let authController: AuthController
    let kycController: KycController
    let walletController: WalletHomeController
    let cardsController: CardsController
    let notificationsController: NotificationsController
    let transferSubmitController: TransferSubmitController

    init() {
        let tokenStorage = TokenStorage()
        let sessionStore = SessionStore()

        let authClient = APIClient(baseURL: AppConfig.apiBaseURL)

        let apiClient = APIClient(
            baseURL: AppConfig.apiBaseURL,
            auth: .init(
                tokenProvider: { tokenStorage.readAccessToken() },
                refresh: {
                    guard let refreshToken = tokenStorage.readRefreshToken() else { return nil }
                    struct RefreshResponse: Decodable, Sendable {
                        let accessToken: String
                        let refreshToken: String
                    }
                    guard let response: RefreshResponse = try? await authClient.send(
                        "/v1/auth/refresh",
                        method: .post,
                        body: ["refresh_token": refreshToken]
                    ) else { return nil }
                    tokenStorage.saveTokens(accessToken: response.accessToken, refreshToken: response.refreshToken)
                    return response.accessToken
                },
                onSessionExpired: {
                    tokenStorage.clear()
                    await sessionStore.setStatus(.unauthenticated)
                }
            )
        )

        let authRepository = AuthRepository(authClient: authClient, client: apiClient)
        let kycRepository = KycRepository(client: apiClient)
        let walletRepository = WalletRepository(client: apiClient)
        let cardRepository = CardRepository(client: apiClient)
        let notificationRepository = NotificationRepository(client: apiClient)
        let transferRepository = TransferRepository(client: apiClient)

        self.sessionStore = sessionStore
        self.authController = AuthController(
            repository: authRepository,
            tokenStorage: tokenStorage,
            sessionStore: sessionStore
        )
        self.kycController = KycController(repository: kycRepository)
        self.walletController = WalletHomeController(repository: walletRepository)
        self.cardsController = CardsController(repository: cardRepository)
        self.notificationsController = NotificationsController(repository: notificationRepository)
        self.transferSubmitController = TransferSubmitController(repository: transferRepository)
    }
}
