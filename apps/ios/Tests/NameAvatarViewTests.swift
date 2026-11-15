import SwiftUI
import XCTest
@testable import MyMoneyRecap

@MainActor
final class NameAvatarViewTests: XCTestCase {
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
            NameAvatarView(model: NameAvatarViewModel(name: name, avatar: avatar) { _, _ in .saved }) {}
        }
        let empty = picture(step("", ""))
        let answered = picture(step("Maya", "otter"))
        XCTAssertNotNil(empty)
        XCTAssertNotEqual(empty, answered, "the saved name and Avatar are drawn, not an empty step")
    }
}
