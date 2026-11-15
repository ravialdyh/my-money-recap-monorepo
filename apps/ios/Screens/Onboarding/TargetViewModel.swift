import Connect
import Foundation
import MoneyRecapContract

@Observable
@MainActor
final class TargetViewModel {
    typealias SaveTarget = (MoneyRecap_V1_BffServiceSaveTargetRequest, Headers) async -> SaveResult

    let salary: MoneyRecap_V1_Money?
    let digits: Int
    var percent: Int
    private(set) var isSaving = false
    private(set) var offline = false
    private var attempt: SaveAttempt<Int?>?
    private let locale: Locale
    private let saveTarget: SaveTarget
    private let saveProfile: CurrencyViewModel.Save

    init(
        salary: MoneyRecap_V1_Money?,
        digits: Int,
        saved: MoneyRecap_V1_SavingTargetSetting? = nil,
        locale: Locale = .autoupdatingCurrent,
        saveTarget: @escaping SaveTarget,
        saveProfile: @escaping CurrencyViewModel.Save
    ) {
        self.salary = salary
        self.digits = digits
        percent = saved.map { Int($0.percent) } ?? 20
        self.locale = locale
        self.saveTarget = saveTarget
        self.saveProfile = saveProfile
    }

    var isBlocked: Bool {
        salary == nil
    }

    var preview: String? {
        salary.flatMap { Self.preview(of: $0, percent: percent, digits: digits, locale: locale) }
    }

    static func preview(of salary: MoneyRecap_V1_Money, percent: Int, digits: Int, locale: Locale) -> String? {
        guard let whole = Decimal(string: salary.amount) else { return nil }
        var share = whole * Decimal(percent) / 100
        var rounded = Decimal()
        NSDecimalRound(&rounded, &share, digits, .plain)
        let exact = rounded.formatted(.number.precision(.fractionLength(digits)).grouping(.never)
            .locale(Locale(identifier: "en_US_POSIX")))
        let printed = { Money.format($0, currency: salary.currencyCode, locale: locale) }
        return "\(printed(exact)) of \(printed(salary.amount)) each month"
    }

    func done() async -> Bool {
        guard !isBlocked, !isSaving else { return false }
        let request = MoneyRecap_V1_BffServiceSaveTargetRequest.with { $0.percent = Int32(percent) }
        return await saving(percent) { await self.saveTarget(request, $0) }
    }

    func skip() async -> Bool {
        guard !isSaving else { return false }
        let request = MoneyRecap_V1_BffServiceSaveProfileRequest.with { $0.onboardingStep = 4 }
        return await saving(nil) { await self.saveProfile(request, $0) }
    }

    private func saving(_ values: Int?, _ send: (Headers) async -> SaveResult) async -> Bool {
        var current = attempt ?? SaveAttempt(values)
        current.update(values)
        attempt = current
        isSaving = true
        defer { isSaving = false }
        let result = await send(current.headers)
        offline = result == .offline
        return result == .saved
    }
}
