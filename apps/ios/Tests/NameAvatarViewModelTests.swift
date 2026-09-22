import Connect
import MoneyRecapContract
import XCTest
@testable import MyMoneyRecap

/// A4 Name and avatar (PRD §6.4): what the field keeps, when Next turns on, and what one save sends.
@MainActor
final class NameAvatarViewModelTests: XCTestCase {
    func testTheFieldKeepsLettersInAnyScriptAndStopsAtThirty() {
        XCTAssertEqual(DisplayName.sanitized("Maya2 😀!"), "Maya ")
        XCTAssertEqual(DisplayName.sanitized("D’Arcy O'Neil-Smith"), "D’Arcy O'Neil-Smith")
        XCTAssertEqual(DisplayName.sanitized("Jose\u{301}"), "Jos\u{E9}", "NFC, as members stores it")
        XCTAssertEqual(DisplayName.sanitized("हिन्दी"), "हिन्दी", "combining marks belong to the letters")
        XCTAssertEqual(DisplayName.sanitized(String(repeating: "a", count: 40)).unicodeScalars.count, 30)
        XCTAssertEqual(DisplayName.stored("  Maya  "), "Maya")
        XCTAssertNil(DisplayName.stored("  2 "), "nothing left once the digit is dropped")
    }

    func testNextNeedsANameAndAnAvatarAndSavesBothWithStepOne() async {
        var sent: [(MoneyRecap_V1_BffServiceSaveProfileRequest, Headers)] = []
        let model = NameAvatarViewModel(name: "", avatar: "") { request, headers in
            sent.append((request, headers))
            return true
        }
        XCTAssertFalse(model.canContinue, "nothing answered yet (SCREEN-STATES A4 empty)")
        model.type(" Maya ")
        XCTAssertFalse(model.canContinue, "no Avatar yet")
        model.choose("owl")
        XCTAssertTrue(model.canContinue)
        let saved = await model.next()
        XCTAssertTrue(saved)
        XCTAssertEqual(sent.count, 1)
        XCTAssertEqual(sent.first?.0.displayName, "Maya")
        XCTAssertEqual(sent.first?.0.avatarKey, "owl")
        XCTAssertEqual(sent.first?.0.onboardingStep, 1)
        XCTAssertFalse(sent.first?.0.hasCurrencyCode ?? true)
        XCTAssertEqual(sent.first?.1["Idempotency-Key"]?.count, 1)
    }

    func testARetryReusesTheKeyAndChangedAnswersGetANewOne() async {
        var keys: [String] = []
        var succeed = false
        let model = NameAvatarViewModel(name: "Test", avatar: "unknown") { _, headers in
            keys.append(headers["Idempotency-Key"]?.first ?? "")
            return succeed
        }
        XCTAssertNil(model.avatar, "a key outside the six is no answer")
        model.choose("fox")
        let failed = await model.next()
        XCTAssertFalse(failed)
        XCTAssertTrue(model.canContinue, "a failed save leaves Next on")
        succeed = true
        _ = await model.next()
        model.choose("frog")
        _ = await model.next()
        XCTAssertEqual(keys.count, 3)
        XCTAssertEqual(keys[0], keys[1], "the retry of the same answers counts once")
        XCTAssertNotEqual(keys[1], keys[2], "a changed Avatar is a new write")
    }
}
