import Connect
import Foundation
import MoneyRecapContract

/// A6 Salary and payday (SCREEN-STATES, PRD §6.6, IOS I6): the take-home amount and the Payday, 25 by default, with
/// the next three paydays. Next saves both with "this month" (spec #92); Skip for now records step 3 alone (A14).
@Observable
@MainActor
final class SalaryViewModel {
    typealias SaveSalary = (MoneyRecap_V1_BffServiceSaveSalaryRequest, Headers) async -> SaveResult

    struct Answers: Equatable {
        let amount: String
        let payday: String
    }

    let currency: String
    /// The Currency's decimals from the server's CLDR data, the most the field takes (spec #92).
    let digits: Int
    var amount: String
    private(set) var payday: String
    private(set) var isSaving = false
    /// The last save could not reach the BFF: the step says so and Next tries again (`NoConnectionView.copy`).
    private(set) var offline = false
    /// The answers being saved, or nil for a skip, so a retry of either keeps its key.
    private var attempt: SaveAttempt<Answers?>?
    private let today: Date
    private let calendar: Calendar
    private let locale: Locale
    private let saveSalary: SaveSalary
    private let saveProfile: CurrencyViewModel.Save

    /// `saved` is step 3's answer already in ledger, filled in again on resume or after going back (spec #92).
    init(
        currency: String,
        digits: Int,
        saved: MoneyRecap_V1_SalarySetting? = nil,
        today: Date = .now,
        calendar: Calendar = .autoupdatingCurrent,
        locale: Locale = .autoupdatingCurrent,
        saveSalary: @escaping SaveSalary,
        saveProfile: @escaping CurrencyViewModel.Save
    ) {
        self.currency = currency
        self.digits = digits
        amount = saved?.amount.amount ?? ""
        payday = saved.map(\.payday).flatMap { $0.isEmpty ? nil : $0 } ?? "25"
        self.today = today
        self.calendar = calendar
        self.locale = locale
        self.saveSalary = saveSalary
        self.saveProfile = saveProfile
    }

    func choose(_ payday: String) {
        self.payday = payday
    }

    /// PRD D51's helper dates, "25 Sep", as the device writes a day and a short month.
    var nextPaydays: [String] {
        let style = Date.FormatStyle(locale: locale, calendar: calendar, timeZone: calendar.timeZone)
            .day().month(.abbreviated)
        return Self.paydays(payday, from: today, calendar: calendar).map { $0.formatted(style) }
    }

    /// The three paydays from `today` on, today included: in each month the Payday, or the month's last day when the
    /// month is shorter or the Payday is "last" (PRD §8.2).
    nonisolated static func paydays(_ payday: String, from today: Date, calendar: Calendar) -> [Date] {
        let day = calendar.startOfDay(for: today)
        let dates = (0 ... 3).compactMap { effective(payday, monthsAfter: $0, today, calendar: calendar) }
        return Array(dates.filter { $0 >= day }.prefix(3))
    }

    /// The effective Payday in the month `offset` months after `date`'s: the day, or the month's last day when the
    /// month is shorter or the Payday is "last" (PRD §8.2).
    nonisolated static func effective(
        _ payday: String,
        monthsAfter offset: Int,
        _ date: Date,
        calendar: Calendar
    ) -> Date? {
        guard let month = calendar.dateInterval(of: .month, for: date)?.start,
              let start = calendar.date(byAdding: .month, value: offset, to: month),
              let length = calendar.range(of: .day, in: .month, for: start)?.count
        else { return nil }
        return calendar.date(byAdding: .day, value: min(Int(payday) ?? length, length) - 1, to: start)
    }

    var canContinue: Bool {
        (Decimal(string: amount) ?? 0) > 0 && !isSaving
    }

    /// True once the Salary is saved. A retry of the same answers sends the same key; changed ones a new key.
    func next() async -> Bool {
        guard canContinue else { return false }
        let sent = amount.hasSuffix(".") ? String(amount.dropLast()) : amount
        let request = MoneyRecap_V1_BffServiceSaveSalaryRequest.with {
            $0.amount = .with {
                $0.amount = sent
                $0.currencyCode = currency
            }
            $0.payday = payday
            $0.appliesFrom = .thisMonth
        }
        return await saving(Answers(amount: sent, payday: payday)) { await self.saveSalary(request, $0) }
    }

    /// True once the skip is recorded: step 3 with no Salary, so step 4 is blocked and Home shows the salary prompt
    /// (D44, D54).
    func skip() async -> Bool {
        guard !isSaving else { return false }
        let request = MoneyRecap_V1_BffServiceSaveProfileRequest.with { $0.onboardingStep = 3 }
        return await saving(nil) { await self.saveProfile(request, $0) }
    }

    private func saving(_ answers: Answers?, _ send: (Headers) async -> SaveResult) async -> Bool {
        var current = attempt ?? SaveAttempt(answers)
        current.update(answers)
        attempt = current
        isSaving = true
        defer { isSaving = false }
        let result = await send(current.headers)
        offline = result == .offline
        return result == .saved
    }
}
