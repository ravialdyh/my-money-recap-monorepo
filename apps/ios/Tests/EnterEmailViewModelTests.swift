import XCTest
@testable import MyMoneyRecap

/// A2 Enter email (SCREEN-STATES): Send disabled until the address is one, the coral helper only after
/// blur or Send, and each failure code with its own copy (PRD §6.2, §6.3).
@MainActor
final class EnterEmailViewModelTests: XCTestCase {
    func testSendStaysDisabledAndSilentWhileTheAddressIsTyped() {
        let model = EnterEmailViewModel(send: { _ in nil })
        for typed in ["m", "ma", "maya", "maya@", "maya@exa"] {
            model.email = typed
            XCTAssertFalse(model.canSend, "\(typed) is not an address yet")
            XCTAssertNil(model.invalidHelper, "the Member is never corrected mid-word")
        }
        model.email = "maya@example.com"
        XCTAssertTrue(model.canSend)
    }

    func testTheHelperAppearsOnBlur() {
        let model = EnterEmailViewModel(send: { _ in nil })
        model.email = "maya@example"
        model.blur()
        XCTAssertEqual(model.invalidHelper, EnterEmailViewModel.invalidCopy, "a domain with no dot is not one")
        model.email = "maya@example.com"
        XCTAssertNil(model.invalidHelper, "the helper goes as soon as the address is one")
    }

    func testTheHelperAppearsOnSendAndNothingIsSent() async {
        var attempts = 0
        let model = EnterEmailViewModel(send: { _ in
            attempts += 1
            return nil
        })
        model.email = "maya at example.com"
        let sent = await model.sendLink()
        XCTAssertNil(sent)
        XCTAssertEqual(attempts, 0, "an invalid address never spends one of the five hourly links")
        XCTAssertEqual(model.invalidHelper, EnterEmailViewModel.invalidCopy)
    }

    func testAValidAddressIsSentTrimmedAndNamesCheckYourInbox() async {
        var seen: String?
        let model = EnterEmailViewModel(send: {
            seen = $0
            return nil
        })
        model.email = "  maya@example.com "
        let sent = await model.sendLink()
        XCTAssertEqual(seen, "maya@example.com")
        XCTAssertEqual(sent, "maya@example.com", "Check your inbox names the address the link went to")
        XCTAssertNil(model.errorMessage)
        XCTAssertFalse(model.isSending)
    }

    func testEachFailureCodeShowsItsOwnCopyAndOnlyTheRateLimitDisablesSend() async {
        for failure in [MagicLinkFailure.expired, .alreadyUsed, .rateLimited, .generic] {
            let model = EnterEmailViewModel(send: { _ in failure })
            model.email = "maya@example.com"
            let sent = await model.sendLink()
            XCTAssertNil(sent, "a refused request never reaches Check your inbox")
            XCTAssertEqual(model.errorMessage, failure.copy)
            XCTAssertEqual(
                model.canSend, failure != .rateLimited,
                "five links in an hour disables Send; the other three leave one tap for a fresh link"
            )
        }
    }

    func testARefusedLinkIsOnScreenBeforeTheMemberTouchesAnything() async {
        let model = EnterEmailViewModel(send: { _ in nil }, initialFailure: .expired)
        XCTAssertEqual(model.errorMessage, MagicLinkFailure.expired.copy)
        model.show(.alreadyUsed)
        XCTAssertEqual(model.errorMessage, MagicLinkFailure.alreadyUsed.copy)
        model.show(nil)
        XCTAssertEqual(model.errorMessage, MagicLinkFailure.alreadyUsed.copy, "nothing new is not a clear")
        model.email = "maya@example.com"
        _ = await model.sendLink()
        XCTAssertNil(model.errorMessage, "the next send clears it")
    }
}
