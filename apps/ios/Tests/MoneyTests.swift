import XCTest
@testable import MyMoneyRecap

/// `Money` prints through the platform formatter (IOS I7); a fixed locale keeps the runner's locale out of the copy.
final class MoneyTests: XCTestCase {
    private let gb = Locale(identifier: "en_GB")

    func testWholeAmountsPrintWithoutAZeroFraction() {
        XCTAssertEqual(Money.format("0", currency: "GBP", locale: gb), "£0")
        XCTAssertEqual(Money.format("2400", currency: "GBP", locale: gb), "£2,400")
        XCTAssertEqual(Money.format("2400.00", currency: "GBP", locale: gb), "£2,400")
    }

    func testFractionsKeepTheCurrencysDecimals() {
        XCTAssertEqual(Money.format("12.40", currency: "GBP", locale: gb), "£12.40")
        XCTAssertEqual(Money.format("15000000", currency: "IDR", locale: Locale(identifier: "id_ID")), "Rp15.000.000")
    }

    /// The amount's own decimals win over the platform's count for the currency (PRD D71): iOS 26.0 and 27.0 give
    /// RSD 0 decimals where ledger's CLDR gives 2, so a stored 100.50 would otherwise print as 101.
    func testAnAmountPrintsExactlyTheDecimalsItCarries() {
        XCTAssert(Money.format("100.50", currency: "RSD", locale: gb).hasSuffix("100.50"))
        XCTAssert(Money.format("12.345", currency: "BHD", locale: gb).hasSuffix("12.345"))
        XCTAssert(Money.format("1500", currency: "HUF", locale: gb).hasSuffix("1,500"))
        XCTAssertEqual(Money.format(".5", currency: "GBP", locale: gb), "£0.5")
    }
}
