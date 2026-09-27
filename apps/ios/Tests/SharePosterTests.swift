import SwiftUI
import XCTest
@testable import MyMoneyRecap

/// The share image (#157): C6's four posters with the canvases' facts, X5 from Budi's recorded rupiah month on an
/// id-ID phone, the five facts it is built from, and Share presenting the system share sheet. Pictures are evidence.
@MainActor
final class SharePosterTests: XCTestCase {
    private static func canvas(_ kind: ArchetypePill.Kind, _ title: String, _ month: String) -> SharePoster {
        let october = month == "October 2026"
        return SharePoster(
            archetype: .init(kind: kind, title: title),
            month: month,
            displayName: "Maya",
            saved: october ? "£310" : "£640",
            spent: october ? "£2,090" : "£1,760"
        )
    }

    private static let four = [
        canvas(.squirrel, "The Squirrel", "September 2026"),
        canvas(.hustler, "The Hustler", "September 2026"),
        canvas(.treatYourself, "Treat-Yourself", "October 2026"),
        canvas(.slowBurner, "The Slow Burner", "October 2026"),
    ]

    private static func x5() async throws -> RecapViewModel {
        let answer = try RecapViewModelTests.recorded("recap-squirrel")
        let model = RecapViewModel(
            month: answer.recap.month,
            fetch: { _ in answer },
            locale: Locale(identifier: "id_ID")
        )
        await model.load()
        return model
    }

    private func keep(_ image: UIImage, _ name: String) throws {
        let attachment = try XCTAttachment(data: XCTUnwrap(image.pngData()), uniformTypeIdentifier: "public.png")
        attachment.name = name
        attachment.lifetime = .keepAlways
        add(attachment)
    }

    private func assertPixels(_ image: UIImage, _ name: String) {
        XCTAssertEqual(image.cgImage?.width, 1080, name)
        XCTAssertEqual(image.cgImage?.height, 1920, name)
    }

    func testThePosterIsBuiltFromTheFiveFactsOnly() async throws {
        let labels = Mirror(reflecting: Self.four[0]).children.map(\.label)
        XCTAssertEqual(labels, ["archetype", "month", "displayName", "saved", "spent"])
        let maya = try await RecapViewModelTests.loaded("recap-maya-september")
        XCTAssertEqual(maya.poster, Self.four[0], "Maya's September is C6 as drawn")
        let budi = try await Self.x5()
        XCTAssertEqual(budi.poster?.saved, "Rp5.600.000")
        XCTAssertEqual(budi.poster?.spent, "Rp9.400.000")
        XCTAssertEqual(budi.poster?.displayName, "Budi")
        XCTAssertEqual(budi.poster?.archetype, .init(kind: .squirrel, title: "The Squirrel"))
        XCTAssertEqual(budi.hero, "Rp15.000.000", "Money In exists and is not a fact of the poster")
    }

    func testNoPosterWhereScreen4HasNoShare() async throws {
        for name in ["recap-no-target", "recap-preview", "recap-sign-up", "recap-no-money-in"] {
            let model = try await RecapViewModelTests.loaded(name)
            XCTAssertNil(model.poster, name)
            model.sharePoster()
            XCTAssertNil(model.shared, name)
        }
    }

    func testTappingShareDrawsThePosterAndPresentsTheShareSheet() async throws {
        let model = try await RecapViewModelTests.loaded("recap-maya-september")
        model.page = 3
        let host = UIHostingController(rootView: RecapPagerView(model: model))
        let scene = try XCTUnwrap(UIApplication.shared.connectedScenes.first as? UIWindowScene)
        let window = UIWindow(windowScene: scene)
        window.rootViewController = host
        window.makeKeyAndVisible()
        defer { window.isHidden = true }
        try await Task.sleep(for: .milliseconds(300))
        XCTAssertNil(host.presentedViewController)
        model.sharePoster()
        let image = try XCTUnwrap(model.shared?.image)
        assertPixels(image, "shared")
        let presented = XCTNSPredicateExpectation(predicate: NSPredicate { _, _ in
            host.presentedViewController.map(Self.holdsShareSheet) ?? false
        }, object: nil)
        await fulfillment(of: [presented], timeout: 5)
        model.shared = nil
        let dismissed = XCTNSPredicateExpectation(predicate: NSPredicate { _, _ in
            host.presentedViewController == nil
        }, object: nil)
        await fulfillment(of: [dismissed], timeout: 5)
    }

    private static func holdsShareSheet(_ controller: UIViewController) -> Bool {
        controller is UIActivityViewController || controller.children.contains(where: holdsShareSheet)
    }

    /// Measured in a browser from the C6 canvases' markup with the bundled fonts: the mascot capped at 620px (the
    /// Slow Burner's is wider, 429px tall), where the lead and the card start, and the title's lines.
    func testTheFourPostersLayOutAsTheCanvasesMeasure() throws {
        let layouts = Self.four.map(\.layout)
        XCTAssertEqual(layouts.map(\.mascot.height), [620, 620, 620, 429])
        XCTAssertEqual(layouts.map(\.lead), [784, 784, 784, 593])
        XCTAssertEqual(layouts.map(\.titleLines), [2, 1, 2, 2])
        XCTAssertEqual(layouts.map { ($0.card.minY * 10).rounded() / 10 }, [1329.6, 1161.6, 1329.6, 1138.6])
        XCTAssertEqual(layouts.map(\.stacked), [false, false, true, true], "£2,090 is 368px at 120px, over 348")
        for (poster, layout) in zip(Self.four, layouts) {
            XCTAssertLessThan(layout.card.maxY, 1920 - 96 - 40.8, poster.archetype.title)
            let image = poster.picture()
            assertPixels(image, poster.archetype.title)
            try keep(image, "c6-" + poster.archetype.title.lowercased().replacingOccurrences(of: " ", with: "-"))
        }
    }

    /// A 30-letter Display Name, the longest allowed, wraps the lead and name lines; the card moves down with them
    /// and still clears the foot, even stacked under the Slow Burner's two-line title.
    func testTheLongestDisplayNameWrapsAndTheCardFollows() throws {
        let long = String(repeating: "W", count: 30)
        let poster = SharePoster(
            archetype: .init(kind: .slowBurner, title: "The Slow Burner"),
            month: "September 2026",
            displayName: long,
            saved: "Rp99.999.999.999",
            spent: "Rp99.999.999.999"
        )
        let layout = poster.layout
        XCTAssertTrue(layout.stacked)
        XCTAssertGreaterThan(layout.card.minY, Self.four[3].layout.card.minY + 2 * 52.8 - 1)
        XCTAssertLessThan(layout.card.maxY, 1920 - 96 - 40.8)
        try keep(poster.picture(), "c6-longest-name")
    }

    /// X5 (PRD D45): "Rp5.600.000" is too wide for its column at 120px, so Saved and Spent stack at 84px.
    func testTheRupiahPosterStacksSavedAndSpent() async throws {
        let model = try await Self.x5()
        let poster = try XCTUnwrap(model.poster)
        let layout = poster.layout
        XCTAssertTrue(layout.stacked)
        XCTAssertLessThan(layout.card.maxY, 1920 - 96 - 40.8)
        let image = poster.picture()
        assertPixels(image, "x5")
        XCTAssertNotEqual(image.pngData(), Self.four[0].picture().pngData())
        try keep(image, "x5-rupiah-poster")
    }
}
