import Foundation

enum LedgerAmount {
    /// The ledger returns amounts at full stored precision (e.g.
    /// "2.50000000" or "500.00000000"); render at 2 decimal places for
    /// display. Goes through `Decimal`, not `Double`, to avoid binary-float
    /// drift — and pins the formatter's locale to `en_US_POSIX`, since
    /// `NumberFormatter` otherwise applies the device locale and would
    /// substitute "," for the decimal point.
    static func formatted(_ raw: String) -> String {
        guard let decimal = Decimal(string: raw) else { return raw }
        let formatter = NumberFormatter()
        formatter.locale = Locale(identifier: "en_US_POSIX")
        formatter.numberStyle = .decimal
        formatter.minimumFractionDigits = 2
        formatter.maximumFractionDigits = 2
        return formatter.string(from: decimal as NSDecimalNumber) ?? raw
    }
}
