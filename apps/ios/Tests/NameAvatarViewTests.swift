import SwiftUI
import XCTest
@testable import MyMoneyRecap

/// Step 1 as drawn (PRD §6.4): the app ships the six Avatars under the contract's keys, the chosen one is ringed, and
/// a Member's answers are drawn in place. The UI tests read the same screen through VoiceOver's labels.
@MainActor
final class NameAvatarViewTests: XCTestCase {
    /// `view` as it draws in a 393 × 852 window, the size of the iPhone the UI tests run on.
    private func picture(_ view: some View) -> Data? {
        let host = UIHostingController(rootView: view)
        let window = UIWindow(frame: CGRect(x: 0, y: 0, width: 393, height: 852))
        window.rootViewController = host
        window.makeKeyAndVisible()
        host.view.layoutIfNeeded()
        return UIGraphicsImageRenderer(bounds: host.view.bounds).image { _ in
            host.view.drawHierarchy(in: host.view.bounds, afterScreenUpdates: true)
        }.pngData()
    }

    func testTheAppShipsTheSixAvatarsUnderTheContractsKeys() {
        XCTAssertEqual(AvatarGrid.keys, ["fox", "otter", "panda", "frog", "owl", "rabbit"], "Profile.avatar_key's list")
        for key in AvatarGrid.keys {
            XCTAssertNotNil(UIImage(named: "ill-avatar-\(key)"), "the \(key) Avatar is in the asset catalog")
        }
    }

    func testTheChosenAvatarIsRinged() {
        let none = picture(AvatarGrid(selected: nil) { _ in })
        let otter = picture(AvatarGrid(selected: "otter") { _ in })
        let fox = picture(AvatarGrid(selected: "fox") { _ in })
        XCTAssertNotNil(none)
        XCTAssertNotEqual(none, otter, "choosing an Avatar draws its ring")
        XCTAssertNotEqual(otter, fox, "the ring follows the choice")
    }

    func testAResumingMembersAnswersAreDrawnInPlace() {
        let step = { (name: String, avatar: String) in
            NameAvatarView(model: NameAvatarViewModel(name: name, avatar: avatar) { _, _ in true }) {}
        }
        let empty = picture(step("", ""))
        let answered = picture(step("Maya", "otter"))
        XCTAssertNotNil(empty)
        XCTAssertNotEqual(empty, answered, "the saved name and Avatar are drawn, not an empty step")
    }
}
