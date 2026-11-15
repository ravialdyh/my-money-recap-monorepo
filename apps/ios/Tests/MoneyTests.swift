import XCTest
@testable import MyMoneyRecap

final class MoneyTests: XCTestCase {
    private let gb = Locale(identifier: "en_GB")

    func testWholeAmountsPrintWithoutAZeroFraction() {
        XCTAssertEqual(Money.format("0", currency: "GBP", locale: gb), "£0")
        XCTAssertEqual(Money.format("2400", currency: "GBP", locale: gb), "£2,400")
        XCTAssertEqual(Money.format("2400.00", currency: "GBP", locale: gb), "£2,400")
    }

    func testFractionsKeepTheCurrencysDecimals() {
        XCTAssertEqual(Money.format("12.40", currency: "GBP", locale: gb), "£12.40")
        XCTAssertEqual(Money.format("150000000", currency: "IDR", locale: Locale(identifier: "id_ID")), "Rp150.000.000")
    }

    func testAnAmountPrintsExactlyTheDecimalsItCarries() {
        XCTAssert(Money.format("100.50", currency: "RSD", locale: gb).hasSuffix("100.50"))
        XCTAssert(Money.format("12.345", currency: "BHD", locale: gb).hasSuffix("12.345"))
        XCTAssert(Money.format("1500", currency: "HUF", locale: gb).hasSuffix("1,500"))
        XCTAssertEqual(Money.format(".5", currency: "GBP", locale: gb), "£0.5")
    }
}
