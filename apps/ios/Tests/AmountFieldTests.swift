import SwiftUI
import XCTest
@testable import MyMoneyRecap

/// The amount field (PRD D57, D71, spec #92): it reads what is typed as the exact decimal the contract carries, in
/// the device's separators, drops a 12th whole digit or a decimal the Currency lacks, and draws the D57 line at 11.
@MainActor
final class AmountFieldTests: XCTestCase {
    private let us = Locale(identifier: "en_US")
    private let indonesia = Locale(identifier: "id_ID")

    func testWhatIsTypedIsReadAsAnExactDecimal() {
        XCTAssertEqual(AmountText.parsed("150,000,000", digits: 0, locale: us), "150000000", "grouping dropped")
        XCTAssertEqual(AmountText.parsed("15.000,5", digits: 2, locale: indonesia), "15000.5", "Indonesian separators")
        XCTAssertEqual(AmountText.parsed("007", digits: 0, locale: us), "7")
        XCTAssertEqual(AmountText.parsed(".", digits: 2, locale: us), "0.")
        XCTAssertEqual(AmountText.parsed("", digits: 2, locale: us), "")
        XCTAssertEqual(AmountText.parsed("12.", digits: 0, locale: us), "12", "a rupiah has no decimals to start")
        XCTAssertEqual(AmountText.parsed("12.345", digits: 3, locale: us), "12.345", "a dinar has three")
    }

    func testAKeystrokeBeyondTheCapOrTheCurrencysDecimalsIsDropped() {
        XCTAssertEqual(AmountText.parsed("12.345", digits: 2, locale: us), "12.34")
        XCTAssertEqual(AmountText.parsed("9,999,999,99999", digits: 0, locale: us), "99999999999", "two keys at once")
        XCTAssertEqual(AmountText.parsed("99,999,999,999", digits: 0, locale: us), "99999999999")
        XCTAssertTrue(AmountText.isFull("99999999999"), "11 digits draw the D57 line")
        XCTAssertTrue(AmountText.isFull("99999999999.5"))
        XCTAssertFalse(AmountText.isFull("9999999999.99"))
    }

    func testTheAmountIsShownAsTheLocaleGroupsIt() {
        XCTAssertEqual(AmountText.shown("150000000", locale: us), "150,000,000")
        XCTAssertEqual(AmountText.shown("15000.5", locale: indonesia), "15.000,5")
        XCTAssertEqual(AmountText.shown("12.", locale: us), "12.", "the point stays while the decimals are typed")
        XCTAssertEqual(AmountText.shown("", locale: us), "")
    }

    func testTheCapDrawsItsLineAndAnotherCurrencyItsSymbol() {
        let field = { (amount: String, currency: String) in
            AmountField(
                label: "Amount, after tax",
                currency: currency,
                digits: 0,
                amount: .constant(amount),
                locale: self.us
            )
            .padding()
            .background(.white)
        }
        XCTAssertNotEqual(picture(field("9999999999", "IDR")), picture(field("99999999999", "IDR")), "the D57 line")
        XCTAssertNotEqual(picture(field("150000000", "IDR")), picture(field("150000000", "JPY")), "IDR against ¥")
    }

    /// #106: the size comes from the printed width in Fredoka bold, not from a count of characters: a long rupiah
    /// Salary steps down from 48 through 36 to 28pt, and two lines of one length but different widths size apart.
    func testTheAmountStepsDownByItsPrintedWidth() {
        XCTAssertEqual(AmountField.font(of: 48).familyName, "Fredoka", "the bundled face is the one measured")
        let short = "Rp 150.000.000"
        let long = "Rp 99.999.999.999"
        let width = { (line: String, size: CGFloat) in
            ceil((line as NSString).size(withAttributes: [.font: AmountField.font(of: size)]).width)
        }
        XCTAssertEqual(AmountField.size(for: short, in: width(short, 48)), 48)
        XCTAssertEqual(AmountField.size(for: long, in: width(long, 48) - 1), 36)
        XCTAssertEqual(AmountField.size(for: long, in: width(long, 36) - 1), 28)
        XCTAssertEqual(AmountField.size(for: long, in: 10), 28, "the smallest step when nothing fits")
        XCTAssertEqual(AmountField.size(for: long, in: 0), 48, "48 before the row is measured")
        let narrow = String(repeating: "1", count: 12)
        let wide = String(repeating: "W", count: 12)
        let room = width(narrow, 48)
        XCTAssertEqual(AmountField.size(for: narrow, in: room), 48)
        XCTAssertLessThan(AmountField.size(for: wide, in: room), 48, "same count, wider print, smaller step")
    }
}
