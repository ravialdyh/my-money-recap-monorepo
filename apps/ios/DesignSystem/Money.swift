import Foundation

/// Money is an exact decimal plus an ISO 4217 code (PRD §7, D71), printed by the platform formatter in the
/// device locale with the currency's own decimals (IOS I7, the Swift twin of `money.ts`). Nothing here knows how
/// many decimals a currency has; a whole amount prints without a zero fraction (README "Money").
enum Money {
    static func format(_ amount: String, currency: String, locale: Locale = .autoupdatingCurrent) -> String {
        let value = Decimal(string: amount) ?? 0
        let style = Decimal.FormatStyle.Currency(code: currency, locale: locale)
        return value.isWhole ? value.formatted(style.precision(.fractionLength(0))) : value.formatted(style)
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
