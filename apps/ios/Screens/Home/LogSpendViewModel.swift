import Connect
import Foundation
import MoneyRecapContract

/// B2 to B4 Log a spend (PRD §6.9) and B7 its edit (§6.10): an amount in the Currency's decimals, a Category, a date in
/// the last 60 days, today by default (D34), and a note. Save answers the fresh Home, or nil after a failure, keeping
/// every value.
@Observable
@MainActor
final class LogSpendViewModel: Identifiable {
    typealias Send = (MoneyRecap_V1_BffServiceLogSpendRequest, Headers) async
        -> Result<MoneyRecap_V1_BffServiceLogSpendResponse, ConnectError>
    typealias Edit = (MoneyRecap_V1_BffServiceEditSpendRequest, Headers) async
        -> Result<MoneyRecap_V1_BffServiceEditSpendResponse, ConnectError>
    typealias Delete = (MoneyRecap_V1_BffServiceDeleteSpendRequest, Headers) async
        -> Result<MoneyRecap_V1_BffServiceDeleteSpendResponse, ConnectError>
    private typealias Write = (Values, Headers) async -> Result<MoneyRecap_V1_HomeView, ConnectError>

    static let failure = "We couldn't save this spend. Try again."
    static let deleteFailure = "We couldn't delete this spend. Try again."
    static let dateHelper = "You can log up to 60 days back, never ahead."
    static let oldDateHelper = "Dates older than 60 days can't be changed."
    /// PRD §6.9's note limit, counted as ledger counts it: code points of the NFC form.
    static let noteLimit = 80

    struct Values: Equatable {
        let amount: String
        let category: Icon.Name
        let date: String
        let note: String
    }

    let currency: String
    let digits: Int
    var amount = ""
    private(set) var category: Icon.Name?
    var date: Date
    private(set) var note = ""
    private(set) var isSaving = false
    private(set) var failed = false
    private(set) var deleteFailed = false
    /// The Spend being edited (B7); nil for Log a spend.
    let spend: MoneyRecap_V1_Spend?
    /// Ledger's earliest date to today, the dates it takes (PRD §8.10, D72); today alone before Home names one.
    let dates: ClosedRange<Date>
    private var attempt: SaveAttempt<Values>?
    private var deletion: SaveAttempt<String>?
    private let calendar: Calendar
    private let locale: Locale
    private let write: Write
    private let delete: Delete?

    /// Log a spend: the Date row reads today until the Member picks another day.
    convenience init(
        currency: String,
        digits: Int,
        earliest: String,
        today: Date = .now,
        calendar: Calendar = .autoupdatingCurrent,
        locale: Locale = .autoupdatingCurrent,
        send: @escaping Send
    ) {
        self.init(
            currency: currency, digits: digits, earliest: earliest, today: today, calendar: calendar, locale: locale,
            spend: nil, delete: nil
        ) { values, headers in
            await send(.with {
                $0.amount = Self.money(values.amount, currency)
                $0.category = Self.contract(values.category)
                $0.date = values.date
                $0.note = values.note
            }, headers).map(\.home)
        }
    }

    /// B7 Edit spend: every field prefilled from `spend`, saved whole by `edit` or deleted by `delete`.
    convenience init(
        editing spend: MoneyRecap_V1_Spend,
        currency: String,
        digits: Int,
        earliest: String,
        today: Date = .now,
        calendar: Calendar = .autoupdatingCurrent,
        locale: Locale = .autoupdatingCurrent,
        edit: @escaping Edit,
        delete: @escaping Delete
    ) {
        self.init(
            currency: currency, digits: digits, earliest: earliest, today: today, calendar: calendar, locale: locale,
            spend: spend, delete: delete
        ) { values, headers in
            await edit(.with {
                $0.spendID = spend.id
                $0.amount = Self.money(values.amount, currency)
                $0.category = Self.contract(values.category)
                $0.date = values.date
                $0.note = values.note
            }, headers).map(\.home)
        }
        amount = spend.amount.amount
        category = SpendListViewModel.glyph(spend.category)
        type(spend.note)
    }

    private init(
        currency: String,
        digits: Int,
        earliest: String,
        today: Date,
        calendar: Calendar,
        locale: Locale,
        spend: MoneyRecap_V1_Spend?,
        delete: Delete?,
        write: @escaping Write
    ) {
        self.currency = currency
        self.digits = digits
        let day = calendar.startOfDay(for: today)
        let iso = Date.ISO8601FormatStyle(timeZone: calendar.timeZone).year().month().day()
        dates = min((try? Date(earliest, strategy: iso)) ?? day, day) ... day
        date = spend.flatMap { try? Date($0.date, strategy: iso) } ?? day
        self.calendar = calendar
        self.locale = locale
        self.spend = spend
        self.delete = delete
        self.write = write
    }

    /// B7 old Spend: its date is outside the days ledger takes, so it stays as it is (PRD §6.10, D72).
    var dateLocked: Bool {
        !dates.contains(date)
    }

    /// The months a write changes, "YYYY-MM": the edited Spend's own and the one its date now names.
    var months: Set<String> {
        Set([spend?.date, day].compactMap { $0.map { String($0.prefix(7)) } })
    }

    private var day: String {
        date.formatted(Date.ISO8601FormatStyle(timeZone: calendar.timeZone).year().month().day())
    }

    func choose(_ category: Icon.Name) {
        self.category = category
    }

    /// Keeps the first 80 code points of the note's NFC form, the most ledger stores.
    func type(_ note: String) {
        let scalars = note.precomposedStringWithCanonicalMapping.unicodeScalars
        self.note = String(String.UnicodeScalarView(scalars.prefix(Self.noteLimit)))
    }

    /// The Date row: "Today, 26 September", or the chosen day alone, as the device writes a day and a month.
    var dateLabel: String {
        let style = Date.FormatStyle(locale: locale, calendar: calendar, timeZone: calendar.timeZone).day().month(.wide)
        let day = date.formatted(style)
        return calendar.isDate(date, inSameDayAs: dates.upperBound) ? "Today, \(day)" : day
    }

    var canSave: Bool {
        (Decimal(string: amount) ?? 0) > 0 && category != nil && !isSaving
    }

    /// The fresh Home once saved. A retry of the same values sends the same key; changed values a new one (IOS I3).
    func save() async -> MoneyRecap_V1_HomeView? {
        guard canSave, let category else { return nil }
        let values = Values(
            amount: amount.hasSuffix(".") ? String(amount.dropLast()) : amount,
            category: category,
            date: day,
            note: note
        )
        var current = attempt ?? SaveAttempt(values)
        current.update(values)
        attempt = current
        isSaving = true
        defer { isSaving = false }
        let result = await write(values, current.headers)
        deleteFailed = false
        guard case let .success(home) = result else {
            failed = true
            return nil
        }
        failed = false
        return home
    }

    /// B8 confirmed: the fresh Home once deleted, or nil after a failure, the Spend kept. A retry sends the same key.
    func deleteSpend() async -> MoneyRecap_V1_HomeView? {
        guard let spend, let delete, !isSaving else { return nil }
        let current = deletion ?? SaveAttempt(spend.id)
        deletion = current
        isSaving = true
        defer { isSaving = false }
        let result = await delete(.with { $0.spendID = spend.id }, current.headers)
        failed = false
        guard case let .success(answer) = result else {
            deleteFailed = true
            return nil
        }
        deleteFailed = false
        return answer.home
    }

    private static func money(_ amount: String, _ currency: String) -> MoneyRecap_V1_Money {
        .with {
            $0.amount = amount
            $0.currencyCode = currency
        }
    }

    /// The contract's Category for a glyph: both follow PRD D20's order.
    static func contract(_ category: Icon.Name) -> MoneyRecap_V1_Category {
        MoneyRecap_V1_Category(rawValue: (Icon.Name.allCases.firstIndex(of: category) ?? 7) + 1) ?? .other
    }
}
