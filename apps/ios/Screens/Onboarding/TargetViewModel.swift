import Connect
import Foundation
import MoneyRecapContract

/// A7 Saving target (SCREEN-STATES, PRD §6.7, IOS I6): a share of the Salary, 20% by default, in steps of 5, with the
/// amount it keeps. Without a Salary it is blocked (D44): Skip for now is the one way on. Done saves; skip records step
/// 4.
@Observable
@MainActor
final class TargetViewModel {
    typealias SaveTarget = (MoneyRecap_V1_BffServiceSaveTargetRequest, Headers) async -> SaveResult

    /// Step 3's Salary; nil when it was skipped, which blocks the step (D44).
    let salary: MoneyRecap_V1_Money?
    /// The Currency's decimals from the server's CLDR data, which the preview rounds to (spec #92).
    let digits: Int
    var percent: Int
    private(set) var isSaving = false
    /// The last save could not reach the BFF: the step says so and Done tries again (`NoConnectionView.copy`).
    private(set) var offline = false
    /// The percent being saved, or nil for a skip, so a retry of either keeps its key.
    private var attempt: SaveAttempt<Int?>?
    private let locale: Locale
    private let saveTarget: SaveTarget
    private let saveProfile: CurrencyViewModel.Save

    /// `saved` is step 4's answer already in ledger, selected again on resume (spec #92).
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

    /// "£480 of £2,400 each month": the one sum the phone does, for display while the knob moves, never stored or sent;
    /// the saved Target's amount is ledger's (ARCHITECTURE A2, IOS I6). Half up to the Currency's decimals, as ledger.
    var preview: String? {
        salary.flatMap { Self.preview(of: $0, percent: percent, digits: digits, locale: locale) }
    }

    /// `preview` for any Salary, which D4 shows too.
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

    /// True once the Target is saved. A retry of the same percent sends the same key; a changed one a new key.
    func done() async -> Bool {
        guard !isBlocked, !isSaving else { return false }
        let request = MoneyRecap_V1_BffServiceSaveTargetRequest.with { $0.percent = Int32(percent) }
        return await saving(percent) { await self.saveTarget(request, $0) }
    }

    /// True once the skip is recorded: step 4 with no Target, which ends setup; Home shows the target prompt (D44).
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
