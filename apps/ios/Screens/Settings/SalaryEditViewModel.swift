import Connect
import Foundation
import MoneyRecapContract

@Observable
@MainActor
final class SalaryEditViewModel: Identifiable {
    typealias Save = (MoneyRecap_V1_BffServiceSaveSalaryRequest, Headers) async
        -> Result<MoneyRecap_V1_BffServiceSaveSalaryResponse, ConnectError>

    struct Answers: Equatable {
        let amount: String
        let payday: String
        let applies: MoneyRecap_V1_SalaryAppliesFrom
    }

    let currency: String
    let digits: Int
    var amount: String
    private(set) var payday: String
    private(set) var isSaving = false
    private(set) var failed = false
    var asking = false
    private let ledger: MoneyRecap_V1_LedgerSettings
    private let today: Date
    private let calendar: Calendar
    private let locale: Locale
    private let save: Save
    private var attempt: SaveAttempt<Answers>?

    init(
        settings: MoneyRecap_V1_SettingsView,
        today: Date = .now,
        calendar: Calendar = .autoupdatingCurrent,
        locale: Locale = .autoupdatingCurrent,
        save: @escaping Save
    ) {
        ledger = settings.ledgerSettings
        currency = settings.profile.currencyCode
        digits = Int(settings.profile.currencyDigits)
        amount = ledger.hasSalary ? ledger.salary.amount.amount : ""
        payday = ledger.hasSalary ? ledger.salary.payday : "25"
        self.today = today
        self.calendar = calendar
        self.locale = locale
        self.save = save
    }

    func choose(_ payday: String) {
        self.payday = payday
    }

    var nextPaydays: [String] {
        SalaryViewModel.paydays(payday, from: today, calendar: calendar).map { $0.formatted(style.day().month()) }
    }

    private var sent: String {
        amount.hasSuffix(".") ? String(amount.dropLast()) : amount
    }

    private var amountChanged: Bool {
        !ledger.hasSalary || Decimal(string: sent) != Decimal(string: ledger.salary.amount.amount)
    }

    private var paydayChanged: Bool {
        ledger.hasSalary && payday != ledger.salary.payday
    }

    var canSave: Bool {
        (Decimal(string: sent) ?? 0) > 0 && (amountChanged || paydayChanged) && !isSaving
    }

    var startsLine: String? {
        guard paydayChanged else { return nil }
        let month = monthName(1)
        return amountChanged ? "Your new salary and payday start in \(month)." : "Your new payday starts in \(month)."
    }

    var waitingLine: String? {
        guard ledger.hasNextSalary, let month = ISODate.month(ledger.nextSalary.effectiveFrom) else { return nil }
        var gmt = Calendar(identifier: .gregorian)
        gmt.timeZone = .gmt
        guard let day = SalaryViewModel.effective(ledger.nextSalary.payday, monthsAfter: 0, month, calendar: gmt)
        else { return nil }
        let next = ledger.nextSalary.amount
        let date = day.formatted(ISODate.style(locale: locale).day().month(.wide))
        let money = Money.format(next.amount, currency: next.currencyCode, locale: locale)
        return "\(money) from your next payday, \(date)."
    }

    var sheetBody: String {
        "You were paid on \(paydayDate(0)). "
            + "A date alone can't tell a correction from a raise, so tell us which this is."
    }

    var correctionRow: String {
        "The \(Money.format(sent, currency: currency, locale: locale)) I was paid on \(paydayDate(0))"
    }

    var correctionDetail: String {
        "Fixes \(monthName(0))"
    }

    var raiseRow: String {
        "My new salary, from \(paydayDate(1))"
    }

    var raiseDetail: String {
        "\(monthName(0)) stays at \(Money.format(ledger.salary.amount.amount, currency: currency, locale: locale))"
    }

    func commit() async -> Bool {
        guard canSave else { return false }
        if ledger.hasSalary, amountChanged, !paydayChanged, ledger.paydayPassedThisMonth {
            asking = true
            return false
        }
        return await send(.unspecified)
    }

    func answer(_ applies: MoneyRecap_V1_SalaryAppliesFrom) async -> Bool {
        asking = false
        return await send(applies)
    }

    private func send(_ applies: MoneyRecap_V1_SalaryAppliesFrom) async -> Bool {
        let answers = Answers(amount: sent, payday: payday, applies: applies)
        var current = attempt ?? SaveAttempt(answers)
        current.update(answers)
        attempt = current
        isSaving = true
        defer { isSaving = false }
        let request = MoneyRecap_V1_BffServiceSaveSalaryRequest.with {
            $0.amount = .with {
                $0.amount = answers.amount
                $0.currencyCode = currency
            }
            $0.payday = answers.payday
            $0.appliesFrom = applies
        }
        switch await save(request, current.headers) {
        case .success:
            failed = false
            return true
        case let .failure(error) where error.code == .failedPrecondition && applies == .unspecified:
            failed = false
            asking = true
            return false
        case .failure:
            failed = true
            return false
        }
    }

    private var style: Date.FormatStyle {
        Date.FormatStyle(locale: locale, calendar: calendar, timeZone: calendar.timeZone)
    }

    private func paydayDate(_ offset: Int) -> String {
        SalaryViewModel.effective(payday, monthsAfter: offset, today, calendar: calendar)
            .map { $0.formatted(style.day().month(.wide)) } ?? ""
    }

    private func monthName(_ offset: Int) -> String {
        calendar.date(byAdding: .month, value: offset, to: today).map { $0.formatted(style.month(.wide)) } ?? ""
    }
}
