import Foundation

/// An exact decimal plus an ISO 4217 code (PRD §7, D71), printed by the platform formatter in the device locale with
/// exactly the decimals ledger rounded it to, not the platform's count for the currency (I7). Whole: no fraction.
enum Money {
    static func format(_ amount: String, currency: String, locale: Locale = .autoupdatingCurrent) -> String {
        let value = Decimal(string: amount) ?? 0
        let point = amount.firstIndex(of: ".")
        let decimals = value.isWhole ? 0 : point.map { amount.distance(from: $0, to: amount.endIndex) - 1 } ?? 0
        let style = Decimal.FormatStyle.Currency(code: currency, locale: locale)
        return value.formatted(style.precision(.fractionLength(decimals)))
    }
}

private extension Decimal {
    var isWhole: Bool {
        var source = self
        var rounded = Decimal()
        NSDecimalRound(&rounded, &source, 0, .plain)
        return rounded == self
    }
}
