import Foundation

/// Amounts are decimal strings straight from the ledger (never `Double`, to
/// avoid floating-point drift) — formatted for display, never parsed and
/// re-computed on-device.
struct WalletBalance: Decodable, Sendable {
    let walletId: String
    let currency: String
    let balance: String
    let availableBalance: String
    let encumberedBalance: String?
}

struct WalletTransaction: Decodable, Sendable, Identifiable {
    let id: String
    let type: String
    let amount: String
    let currency: String
    let direction: String
    let status: String
    let createdAt: Date
    let counterparty: String?
    let memo: String?
    let referenceId: String?

    var isCredit: Bool { direction == "credit" }
}

struct WalletTransactionPage: Decodable, Sendable {
    let transactions: [WalletTransaction]
    let nextCursor: String?
}
