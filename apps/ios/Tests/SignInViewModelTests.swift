import XCTest
@testable import MyMoneyRecap

@MainActor
final class SignInViewModelTests: XCTestCase {
    func testSuccessLeavesNoError() async {
        let model = SignInViewModel(signIn: {})
        await model.continueWithEmail()
        XCTAssertNil(model.errorMessage)
        XCTAssertFalse(model.isLoading)
    }

    func testCancelOrFailureShowsInlineErrorAndReenables() async {
        let model = SignInViewModel(signIn: { throw CancellationError() })
        await model.continueWithEmail()
        XCTAssertEqual(model.errorMessage, SignInViewModel.errorCopy)
        XCTAssertFalse(model.isLoading, "buttons are re-enabled after a failure")
    }

    func testNextTapClearsErrorAndShowsLoading() async {
        var attempts = 0
        var seen: (error: String?, loading: Bool)?
        var model: SignInViewModel!
        model = SignInViewModel(signIn: {
            attempts += 1
            if attempts == 1 {
                throw CancellationError()
            }
            seen = (model.errorMessage, model.isLoading)
        })
        await model.continueWithEmail()
        XCTAssertNotNil(model.errorMessage)
        await model.continueWithEmail()
        XCTAssertNil(seen?.error, "the error is cleared before the next round-trip starts")
        XCTAssertEqual(seen?.loading, true)
        XCTAssertNil(model.errorMessage)
    }
}
