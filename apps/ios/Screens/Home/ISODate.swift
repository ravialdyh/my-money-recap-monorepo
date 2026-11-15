import Foundation

enum ISODate {
    static func style(locale: Locale) -> Date.FormatStyle {
        Date.FormatStyle(locale: locale, timeZone: .gmt)
    }

    static func day(_ string: String) -> Date? {
        try? Date(string, strategy: .iso8601.year().month().day())
    }

    static func month(_ string: String) -> Date? {
        day(string + "-01")
    }
}
