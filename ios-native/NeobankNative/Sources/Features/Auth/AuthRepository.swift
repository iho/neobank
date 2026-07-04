import Foundation

/// Wraps the two auth-adjacent clients: `authClient` for the unauthenticated
/// `/v1/auth/login` and `/v1/auth/register` endpoints, `client` for calls that
/// need (and may transparently refresh) a bearer token.
struct AuthRepository: Sendable {
    let authClient: APIClient
    let client: APIClient

    func login(email: String, password: String) async throws -> AuthTokens {
        try await authClient.send(
            "/v1/auth/login",
            method: .post,
            body: ["email": email, "password": password]
        )
    }

    func register(email: String, password: String, phone: String?, inviteCode: String?) async throws -> AuthTokens {
        try await authClient.send(
            "/v1/auth/register",
            method: .post,
            body: [
                "email": email,
                "password": password,
                "phone": phone?.isEmpty == false ? phone : nil,
                "invite_code": inviteCode?.isEmpty == false ? inviteCode : nil,
            ]
        )
    }

    func profile() async throws -> Profile {
        try await client.send("/v1/me")
    }

    func changePassword(current: String, new: String) async throws {
        let _: EmptyResponse = try await client.send(
            "/v1/auth/change-password",
            method: .post,
            body: ["current_password": current, "new_password": new]
        )
    }
}
