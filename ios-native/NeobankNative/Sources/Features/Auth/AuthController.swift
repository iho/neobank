import Foundation
import Observation

/// Drives the auth lifecycle: reading the persisted session on launch, and
/// performing login/register/logout. Screens own their own submitting/error
/// state; this just owns the actions and the resulting session transition.
@MainActor
@Observable
final class AuthController {
    private let repository: AuthRepository
    private let tokenStorage: TokenStorage
    private let sessionStore: SessionStore

    init(repository: AuthRepository, tokenStorage: TokenStorage, sessionStore: SessionStore) {
        self.repository = repository
        self.tokenStorage = tokenStorage
        self.sessionStore = sessionStore
    }

    func bootstrap() {
        let hasToken = tokenStorage.readAccessToken() != nil
        sessionStore.setStatus(hasToken ? .authenticated : .unauthenticated)
    }

    func login(email: String, password: String) async throws {
        let tokens = try await repository.login(email: email, password: password)
        tokenStorage.saveTokens(accessToken: tokens.accessToken, refreshToken: tokens.refreshToken)
        sessionStore.setStatus(.authenticated)
    }

    func register(email: String, password: String, phone: String?, inviteCode: String?) async throws {
        let tokens = try await repository.register(
            email: email,
            password: password,
            phone: phone,
            inviteCode: inviteCode
        )
        tokenStorage.saveTokens(accessToken: tokens.accessToken, refreshToken: tokens.refreshToken)
        sessionStore.setStatus(.authenticated)
    }

    func logout() {
        tokenStorage.clear()
        sessionStore.setStatus(.unauthenticated)
    }
}

#if DEBUG
extension AuthController {
    static var preview: AuthController {
        let client = APIClient(baseURL: AppConfig.apiBaseURL)
        return AuthController(
            repository: AuthRepository(authClient: client, client: client),
            tokenStorage: TokenStorage(),
            sessionStore: SessionStore()
        )
    }
}
#endif
