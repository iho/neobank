import Foundation

/// Composition root: wires the unauthenticated auth client, the token-bearing
/// client (with refresh-on-401), and the objects built on top of them.
@MainActor
struct AppEnvironment {
    let sessionStore: SessionStore
    let authController: AuthController
    let kycController: KycController

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

        self.sessionStore = sessionStore
        self.authController = AuthController(
            repository: authRepository,
            tokenStorage: tokenStorage,
            sessionStore: sessionStore
        )
        self.kycController = KycController(repository: kycRepository)
    }
}
