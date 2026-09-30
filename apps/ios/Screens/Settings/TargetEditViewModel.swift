import Connect
import Foundation
import MoneyRecapContract

/// D4 Saving target (PRD §6.16, D65): setup's slider over the last Target saved, 20% with none, and ledger's month a
/// save starts in. Without a Salary it is blocked and points to D3 (D44).
@Observable
@MainActor
final class TargetEditViewModel: Identifiable {
    typealias Save = (MoneyRecap_V1_BffServiceSaveTargetRequest, Headers) async
        -> Result<MoneyRecap_V1_BffServiceSaveTargetResponse, ConnectError>

    /// The Salary in force; nil blocks the screen (D44).
    let salary: MoneyRecap_V1_Money?
    let digits: Int
    var percent: Int
    private(set) var isSaving = false
    /// The last save failed: the slider stays where it was, over `SettingsViewModel.saveFailure`.
    private(set) var failed = false
    private let starts: String
    private let today: Date
    private let calendar: Calendar
    private let locale: Locale
    private let save: Save
    private var attempt: SaveAttempt<Int>?

    init(
        settings: MoneyRecap_V1_SettingsView,
        today: Date = .now,
        calendar: Calendar = .autoupdatingCurrent,
        locale: Locale = .autoupdatingCurrent,
        save: @escaping Save
    ) {
        let ledger = settings.ledgerSettings
        salary = ledger.hasSalary ? ledger.salary.amount : nil
        digits = Int(settings.profile.currencyDigits)
        let saved = ledger.hasNextSavingTarget ? ledger.nextSavingTarget
            : ledger.hasSavingTarget ? ledger.savingTarget : nil
        percent = saved.map { Int($0.percent) } ?? 20
        starts = ledger.newSavingTargetFrom
        self.today = today
        self.calendar = calendar
        self.locale = locale
        self.save = save
    }

    var isBlocked: Bool {
        salary == nil
    }

    /// Setup's "{amount} of {salary} each month", the phone's one sum (ARCHITECTURE A2).
    var preview: String? {
        salary.flatMap { TargetViewModel.preview(of: $0, percent: percent, digits: digits, locale: locale) }
    }

    /// "Your new target starts in October." only when ledger says a save now starts after this month (D65).
    var startsLine: String? {
        let now = calendar.dateComponents([.year, .month], from: today)
        let current = String(format: "%04d-%02d", now.year ?? 0, now.month ?? 0)
        guard starts > current, let month = ISODate.month(starts) else { return nil }
        return "Your new target starts in \(month.formatted(ISODate.style(locale: locale).month(.wide)))."
    }

    var canSave: Bool {
        !isBlocked && !isSaving
    }

    /// True once saved. A retry of the same percent sends the same key; a failure keeps the slider where it was.
    func commit() async -> Bool {
        guard canSave else { return false }
        var current = attempt ?? SaveAttempt(percent)
        current.update(percent)
        attempt = current
        isSaving = true
        defer { isSaving = false }
        let request = MoneyRecap_V1_BffServiceSaveTargetRequest.with { $0.percent = Int32(percent) }
        switch await save(request, current.headers) {
        case .success:
            failed = false
            return true
        case .failure:
            failed = true
            return false
        }
    }
}
