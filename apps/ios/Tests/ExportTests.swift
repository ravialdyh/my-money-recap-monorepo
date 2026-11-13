import Connect
import MoneyRecapContract
import XCTest
@testable import MyMoneyRecap

/// D9 Export spends (SCREEN-STATES D9): the card naming the address, `Sent ✓` at 40% for 60 seconds, the refused
/// address named with no link, and any other failure retried under the same key.
@MainActor
final class ExportTests: XCTestCase {
    /// Holds the 60 seconds open until the test lets them pass.
    @MainActor
    private final class Clock {
        var waited: [Duration] = []
        private var release: CheckedContinuation<Void, Never>?

        func wait(_ duration: Duration) async {
            waited.append(duration)
            await withCheckedContinuation { release = $0 }
        }

        /// Waits until the 60 seconds have begun, then lets them pass.
        func pass() async {
            while release == nil {
                await Task.yield()
            }
            release?.resume()
            release = nil
        }

        /// Yields until the 60 seconds have begun.
        func started(_ count: Int) async {
            while waited.count < count {
                await Task.yield()
            }
        }
    }

    private func model(
        _ answers: [Result<MoneyRecap_V1_BffServiceRequestExportResponse, ConnectError>],
        keys: @escaping (String) -> Void = { _ in },
        clock: Clock? = nil
    ) -> ExportViewModel {
        var answers = answers
        return ExportViewModel(email: "test@example.com", send: { headers in
            keys(headers["Idempotency-Key"]?.first ?? "")
            return answers.removeFirst()
        }, wait: (clock ?? Clock()).wait)
    }

    func testSentReadsSentForSixtySecondsThenSendItAgain() async {
        let clock = Clock()
        let model = model([.success(.init())], clock: clock)
        XCTAssertEqual(
            model.helper,
            "It goes to test@example.com. Every spend you've ever logged, with its date, category, amount and note."
        )
        XCTAssertEqual(model.button, "Send it")
        XCTAssertTrue(model.canSend)
        let sending = Task { await model.commit() }
        await clock.started(1)
        XCTAssertEqual(model.button, "Sent ✓")
        XCTAssertFalse(model.canSend, "dimmed to 40%, so a double tap cannot send twice (D67)")
        XCTAssertNil(model.error)
        XCTAssertEqual(clock.waited, [.seconds(60)])
        await clock.pass()
        await sending.value
        XCTAssertEqual(model.button, "Send it")
        XCTAssertTrue(model.canSend)
    }

    /// notify's refusal as the API passes it on: failed_precondition carrying EmailBounced.
    private static let bounced = ConnectError(code: .failedPrecondition, message: nil, details: [
        .init(
            type: MoneyRecap_V1_EmailBounced.protoMessageName,
            payload: try? MoneyRecap_V1_EmailBounced().serializedData()
        ),
    ])

    func testARefusedAddressIsNamedWithNoWayToResend() async {
        let model = model([.failure(Self.bounced)])
        await model.commit()
        XCTAssertEqual(model.error, "We can't reach test@example.com.")
        XCTAssertEqual(model.helper, "Every spend you've ever logged, with its date, category, amount and note.")
        XCTAssertFalse(model.canSend, "Profile cannot change the address, so there is nothing to retry")
        let plain = self.model([.failure(ConnectError(code: .failedPrecondition, message: nil))])
        await plain.commit()
        XCTAssertEqual(plain.error, "We couldn't send it. Try again.", "only EmailBounced names the address")
    }

    func testAnyOtherFailureRetriesUnderTheSameKey() async {
        var keys: [String] = []
        let clock = Clock()
        let model = model(
            [.failure(ConnectError(code: .unavailable, message: nil)), .success(.init()), .success(.init())],
            keys: { keys.append($0) }, clock: clock
        )
        await model.commit()
        XCTAssertEqual(model.error, "We couldn't send it. Try again.")
        XCTAssertTrue(model.canSend)
        let retry = Task { await model.commit() }
        await clock.pass()
        await retry.value
        let again = Task { await model.commit() }
        await clock.pass()
        await again.value
        XCTAssertEqual(keys.count, 3)
        XCTAssertEqual(keys[0], keys[1], "the retry repeats the failed attempt's key (A14)")
        XCTAssertNotEqual(keys[1], keys[2], "a new export after the 60 seconds is a new key")
    }

    func testSettingsOpensExportAtTheSignedInAddress() async throws {
        let filled = try SettingsViewModelTests.recorded("settings-filled")
        let settings = try SettingsViewModelTests.model({ filled }, context: SettingsViewModelTests.context())
        settings.openExport()
        XCTAssertNil(settings.export, "no Settings yet: nothing to name")
        await settings.refresh()
        settings.openExport()
        XCTAssertEqual(settings.export?.email, filled.profile.email)
    }

    func testTheStatesDrawApartAndScrollSideways() async throws {
        let ready = model([])
        let refused = model([.failure(Self.bounced)])
        await refused.commit()
        let failed = model([.failure(ConnectError(code: .unavailable, message: nil))])
        await failed.commit()
        let sideways = CGSize(width: 852, height: 393)
        let drawn = try XCTUnwrap(picture(ExportView(model: ready)))
        let refusal = picture(ExportView(model: refused))
        XCTAssertNotEqual(drawn, refusal, "the refusal line in the card")
        XCTAssertNotEqual(drawn, picture(ExportView(model: failed)), "the retry line in the card")
        XCTAssertGreaterThan(overflow(ExportView(model: failed), size: sideways), 0)
        try keep(refusal, "d9-refused")
        try keep(picture(ExportView(model: refused), size: sideways), "d9-refused-sideways")
        try keep(picture(ExportView(model: failed)), "d9-send-failed")
    }

    private func keep(_ png: Data?, _ name: String) throws {
        let attachment = try XCTAttachment(data: XCTUnwrap(png), uniformTypeIdentifier: "public.png")
        attachment.name = name
        attachment.lifetime = .keepAlways
        add(attachment)
    }
}
