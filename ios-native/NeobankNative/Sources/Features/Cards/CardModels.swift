import Foundation

struct BankCard: Decodable, Sendable, Identifiable, Equatable {
    let id: String
    let userId: String
    let walletId: String
    let lastFour: String
    let status: String
    let expiryMonth: Int
    let expiryYear: Int
    let onlineOnly: Bool
    let dailyLimit: String?

    var isFrozen: Bool { status == "frozen" }

    /// Plain (non-localized) "MM/YYYY". Built as a `String`, not interpolated
    /// straight into a `Text` literal — `Text("...\(expiryYear)")` binds to
    /// the `LocalizedStringKey` overload, which runs raw `Int` interpolations
    /// through locale-aware number formatting and inserts a grouping
    /// separator (e.g. "2 029" under locales that group with a space).
    var expiry: String {
        String(format: "%02d/%d", expiryMonth, expiryYear)
    }
}
