import Connect
import Foundation
import MoneyRecapContract

@Observable
@MainActor
final class SalaryViewModel {
    typealias SaveSalary = (MoneyRecap_V1_BffServiceSaveSalaryRequest, Headers) async -> SaveResult

    struct Answers: Equatable {
        let amount: String
        let payday: String
    }

    let currency: String
    let digits: Int
    var amount: String
    private(set) var payday: String
    private(set) var isSaving = false
    private(set) var offline = false
    private var attempt: SaveAttempt<Answers?>?
    private let today: Date
    private let calendar: Calendar
    private let locale: Locale
    private let saveSalary: SaveSalary
    private let saveProfile: CurrencyViewModel.Save

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

    var nextPaydays: [String] {
        let style = Date.FormatStyle(locale: locale, calendar: calendar, timeZone: calendar.timeZone)
            .day().month(.abbreviated)
        return Self.paydays(payday, from: today, calendar: calendar).map { $0.formatted(style) }
    }

    nonisolated static func paydays(_ payday: String, from today: Date, calendar: Calendar) -> [Date] {
        let day = calendar.startOfDay(for: today)
        let dates = (0 ... 3).compactMap { effective(payday, monthsAfter: $0, today, calendar: calendar) }
        return Array(dates.filter { $0 >= day }.prefix(3))
    }

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
