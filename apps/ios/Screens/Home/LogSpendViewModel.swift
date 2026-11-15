import Connect
import Foundation
import MoneyRecapContract

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
    let spend: MoneyRecap_V1_Spend?
    let dates: ClosedRange<Date>
    private var attempt: SaveAttempt<Values>?
    private var deletion: SaveAttempt<String>?
    private let calendar: Calendar
    private let locale: Locale
    private let write: Write
    private let delete: Delete?
    private let connection: Connection?

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
            spend: nil, delete: nil, connection: nil
        ) { values, headers in
            await send(.with {
                $0.amount = Self.money(values.amount, currency)
                $0.category = Self.contract(values.category)
                $0.date = values.date
                $0.note = values.note
            }, headers).map(\.home)
        }
    }

    convenience init(
        editing spend: MoneyRecap_V1_Spend,
        currency: String,
        digits: Int,
        earliest: String,
        today: Date = .now,
        calendar: Calendar = .autoupdatingCurrent,
        locale: Locale = .autoupdatingCurrent,
        connection: Connection? = nil,
        edit: @escaping Edit,
        delete: @escaping Delete
    ) {
        self.init(
            currency: currency, digits: digits, earliest: earliest, today: today, calendar: calendar, locale: locale,
            spend: spend, delete: delete, connection: connection ?? .shared
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
        connection: Connection?,
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
        self.connection = connection
        self.write = write
    }

    var dateLocked: Bool {
        !dates.contains(date)
    }

    var months: Set<String> {
        Set([spend?.date, day].compactMap { $0.map { String($0.prefix(7)) } })
    }

    private var day: String {
        date.formatted(Date.ISO8601FormatStyle(timeZone: calendar.timeZone).year().month().day())
    }

    func choose(_ category: Icon.Name) {
        self.category = category
    }

    func type(_ note: String) {
        let scalars = note.precomposedStringWithCanonicalMapping.unicodeScalars
        self.note = String(String.UnicodeScalarView(scalars.prefix(Self.noteLimit)))
    }

    var dateLabel: String {
        let style = Date.FormatStyle(locale: locale, calendar: calendar, timeZone: calendar.timeZone).day().month(.wide)
        let day = date.formatted(style)
        return calendar.isDate(date, inSameDayAs: dates.upperBound) ? "Today, \(day)" : day
    }

    var canSave: Bool {
        (Decimal(string: amount) ?? 0) > 0 && category != nil && !isSaving && !offline
    }

    var offline: Bool {
        !(connection?.online ?? true)
    }

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

    func deleteSpend() async -> MoneyRecap_V1_HomeView? {
        guard let spend, let delete, !isSaving, !offline else { return nil }
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

    static func contract(_ category: Icon.Name) -> MoneyRecap_V1_Category {
        MoneyRecap_V1_Category(rawValue: (Icon.Name.allCases.firstIndex(of: category) ?? 7) + 1) ?? .other
    }
}
