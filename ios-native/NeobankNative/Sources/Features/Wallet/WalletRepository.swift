import Foundation

struct WalletRepository: Sendable {
    let client: APIClient

    func balance(currency: String = "USD") async throws -> WalletBalance {
        try await client.send("/v1/wallet", query: ["currency": currency])
    }

    func transactions(cursor: String? = nil, limit: Int = 20) async throws -> WalletTransactionPage {
        try await client.send(
            "/v1/wallet/transactions",
            query: ["limit": String(limit), "cursor": cursor]
        )
    }
}
