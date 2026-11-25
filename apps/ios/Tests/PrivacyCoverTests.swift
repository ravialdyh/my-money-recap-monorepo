import UIKit
import XCTest
@testable import MyMoneyRecap

@MainActor
final class PrivacyCoverTests: XCTestCase {
    func testTheCoverSitsAboveEverySheetUntilItIsTakenDown() throws {
        let scene = try XCTUnwrap(UIApplication.shared.connectedScenes.compactMap { $0 as? UIWindowScene }.first)
        func covers() -> [UIWindow] {
            scene.windows.filter { !$0.isHidden && $0.windowLevel > .alert }
        }
        PrivacyCover.show()
        PrivacyCover.show()
        XCTAssertEqual(covers().count, 1, "one cover, above alerts and so above every sheet")
        PrivacyCover.hide()
        XCTAssertTrue(covers().isEmpty)
    }
}
