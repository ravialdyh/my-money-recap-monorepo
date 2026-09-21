import XCTest
@testable import MyMoneyRecap

@MainActor
final class SignInViewModelTests: XCTestCase {
    func testSuccessLeavesNoError() async {
        let model = SignInViewModel(signIn: { _ in })
        await model.tap(.email)
        XCTAssertNil(model.errorMessage)
        XCTAssertNil(model.tapped)
    }

    func testTheTappedButtonSpinsWhileTheOtherTwoDim() async {
        for button in [SignInButton.apple, .google, .email] {
            var model: SignInViewModel!
            var sent: SignInButton?
            var seen: (tapped: SignInButton?, dimmed: [Bool])?
            model = SignInViewModel(signIn: { pressed in
                sent = pressed
                seen = (model.tapped, [model.dims(.apple), model.dims(.google), model.dims(.email)])
            })
            await model.tap(button)
            XCTAssertEqual(sent, button, "the round-trip is for the button that was tapped")
            XCTAssertEqual(seen?.tapped, button, "\(button) shows the spinner while its sheet is open")
            XCTAssertEqual(
                seen?.dimmed,
                [button != .apple, button != .google, button != .email],
                "\(button) stays at full strength and the other two sit at 40%"
            )
        }
    }

    func testCancelShowsInlineErrorWithAllThreeLiveAgain() async {
        let model = SignInViewModel(signIn: { _ in throw CancellationError() })
        await model.tap(.google)
        XCTAssertEqual(model.errorMessage, SignInViewModel.errorCopy)
        XCTAssertNil(model.tapped, "no spinner after a cancel")
        XCTAssertEqual(
            [model.dims(.apple), model.dims(.google), model.dims(.email)],
            [false, false, false],
            "a mis-tap costs nothing: all three buttons are live again"
        )
    }

    func testATapWhileASheetIsOpenIsIgnored() async {
        var sent: [SignInButton] = []
        var model: SignInViewModel!
        model = SignInViewModel(signIn: { button in
            sent.append(button)
            await model.tap(.google)
        })
        await model.tap(.apple)
        XCTAssertEqual(sent, [.apple], "one web sheet at a time")
    }

    func testNextTapClearsErrorAndShowsLoading() async {
        var attempts = 0
        var seen: (error: String?, tapped: SignInButton?)?
        var model: SignInViewModel!
        model = SignInViewModel(signIn: { _ in
            attempts += 1
            if attempts == 1 {
                throw CancellationError()
            }
            seen = (model.errorMessage, model.tapped)
        })
        await model.tap(.apple)
        XCTAssertNotNil(model.errorMessage)
        await model.tap(.apple)
        XCTAssertNil(seen?.error, "the error is cleared before the next round-trip starts")
        XCTAssertEqual(seen?.tapped, .apple)
        XCTAssertNil(model.errorMessage)
    }

    func testAppleAndGoogleNameTheirProviderAndEmailNamesNone() {
        XCTAssertEqual(SignInButton.apple.identityProvider, .apple)
        XCTAssertEqual(SignInButton.google.identityProvider, .google)
        XCTAssertNil(SignInButton.email.identityProvider, "no hint: Keycloak's own page, the debug seam")
    }
}
