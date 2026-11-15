import Connect
import Foundation
import MoneyRecapContract

@Observable
@MainActor
final class TargetEditViewModel: Identifiable {
    typealias Save = (MoneyRecap_V1_BffServiceSaveTargetRequest, Headers) async
        -> Result<MoneyRecap_V1_BffServiceSaveTargetResponse, ConnectError>

    let salary: MoneyRecap_V1_Money?
    let digits: Int
    var percent: Int
    private(set) var isSaving = false
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

    var preview: String? {
        salary.flatMap { TargetViewModel.preview(of: $0, percent: percent, digits: digits, locale: locale) }
    }

    var startsLine: String? {
        let now = calendar.dateComponents([.year, .month], from: today)
        let current = String(format: "%04d-%02d", now.year ?? 0, now.month ?? 0)
        guard starts > current, let month = ISODate.month(starts) else { return nil }
        return "Your new target starts in \(month.formatted(ISODate.style(locale: locale).month(.wide)))."
    }

    var canSave: Bool {
        !isBlocked && !isSaving
    }

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
