import Connect
import MoneyRecapContract
import XCTest
@testable import MyMoneyRecap

/// D10 and D11 Delete account (SCREEN-STATES D10, D11): only exactly `DELETE` asks, a granted request signs out, a
/// failure keeps the Member signed in with its line and retries under the same key, and both screens scroll sideways.
@MainActor
final class DeleteAccountTests: XCTestCase {
    typealias Answer = Result<MoneyRecap_V1_BffServiceRequestDeletionResponse, ConnectError>

    func testOnlyExactlyDELETEAsks() async {
        var asked = 0
        let model = DeleteAccountViewModel(request: { _ in
            asked += 1
            return .success(.init())
        }, signOut: {})
        for typed in ["", "delete", "DELET", "DELETE ", " DELETE"] {
            model.typed = typed
            XCTAssertFalse(model.canDelete, "\(typed.debugDescription) is not DELETE")
            await model.delete()
        }
        XCTAssertEqual(asked, 0)
        model.typed = "DELETE"
        XCTAssertTrue(model.canDelete)
    }

    func testAGrantedRequestSignsOut() async {
        var signedOut = false
        let model = DeleteAccountViewModel(request: { _ in .success(.init()) }, signOut: { signedOut = true })
        model.typed = "DELETE"
        await model.delete()
        XCTAssertTrue(signedOut, "signs out as Log out does")
        XCTAssertFalse(model.failed)
    }

    func testAFailureKeepsTheMemberSignedInAndRetriesUnderTheSameKey() async {
        var answers: [Answer] = [.failure(ConnectError(code: .unavailable, message: nil)), .success(.init())]
        var keys: [String] = [], signedOut = false
        let model = DeleteAccountViewModel(request: { headers in
            keys.append(headers["Idempotency-Key"]?.first ?? "")
            return answers.removeFirst()
        }, signOut: { signedOut = true })
        model.typed = "DELETE"
        await model.delete()
        XCTAssertTrue(model.failed)
        XCTAssertFalse(signedOut, "a failure keeps the Member signed in")
        XCTAssertEqual(DeleteAccountViewModel.failedCopy, "We couldn't delete your account. Try again.")
        await model.delete()
        XCTAssertTrue(signedOut)
        XCTAssertEqual(keys.count, 2)
        XCTAssertEqual(keys[0], keys[1], "the retry repeats the failed attempt's key (A14)")
    }

    func testSettingsSignsOutThroughOneLogOut() async throws {
        var loggedOut = 0
        let settings = try SettingsViewModel(
            calls: SettingsCalls(
                fetch: { try SettingsViewModelTests.recorded("settings-filled") }, saveProfile: { _, _ in .saved },
                requestDeletion: { _ in .success(.init()) }, logOut: { loggedOut += 1 }
            ), context: SettingsViewModelTests.context()
        )
        settings.openDeleteAccount()
        let delete = try XCTUnwrap(settings.deleteAccount, "the row opens D10")
        delete.typed = "DELETE"
        await delete.delete()
        XCTAssertEqual(loggedOut, 1, "a granted deletion signs out as Log out does")
        await settings.logOutNow()
        XCTAssertEqual(loggedOut, 2, "D12's Log out")
    }

    func testTheScreensDrawAndScrollSideways() async throws {
        let model = DeleteAccountViewModel(
            request: { _ in .failure(ConnectError(code: .unavailable, message: nil)) },
            signOut: {}
        )
        let sideways = CGSize(width: 852, height: 393)
        let warning = try XCTUnwrap(picture(DeleteAccountView(model: model)))
        let empty = try XCTUnwrap(picture(DeleteConfirmView(model: model)))
        model.typed = "DELETE"
        await model.delete()
        let failed = try XCTUnwrap(picture(DeleteConfirmView(model: model)))
        XCTAssertNotEqual(empty, failed, "the failure line and the live button")
        XCTAssertGreaterThan(overflow(DeleteAccountView(model: model), size: sideways), 0)
        XCTAssertGreaterThan(overflow(DeleteConfirmView(model: model), size: sideways), 0)
        try keep(warning, "d10-delete-account")
        try keep(picture(DeleteAccountView(model: model), size: sideways), "d10-delete-account-sideways")
        try keep(failed, "d11-delete-failed")
        try keep(picture(DeleteConfirmView(model: model), size: sideways), "d11-delete-failed-sideways")
    }

    private func keep(_ png: Data?, _ name: String) throws {
        let attachment = try XCTAttachment(data: XCTUnwrap(png), uniformTypeIdentifier: "public.png")
        attachment.name = name
        attachment.lifetime = .keepAlways
        add(attachment)
    }
}
