import Foundation

struct CardRepository: Sendable {
    let client: APIClient

    private struct CardsResponse: Decodable, Sendable {
        let cards: [BankCard]
    }

    func list() async throws -> [BankCard] {
        let response: CardsResponse = try await client.send("/v1/cards")
        return response.cards
    }

    func issue(cardholderName: String, dailyLimit: String?, onlineOnly: Bool?) async throws -> BankCard {
        try await client.send(
            "/v1/cards",
            method: .post,
            body: [
                "cardholder_name": cardholderName,
                "daily_limit": dailyLimit?.isEmpty == false ? dailyLimit : nil,
                "online_only": onlineOnly,
            ]
        )
    }

    func freeze(_ id: String) async throws -> BankCard {
        try await client.send("/v1/cards/\(id)/freeze", method: .post)
    }

    func unfreeze(_ id: String) async throws -> BankCard {
        try await client.send("/v1/cards/\(id)/unfreeze", method: .post)
    }
}
