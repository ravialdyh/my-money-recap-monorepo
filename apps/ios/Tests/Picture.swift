import SwiftUI
import XCTest

extension XCTestCase {
    /// `view` as it draws in a 393 × 852 window, the size of the iPhone the UI tests run on. SwiftUI's accessibility
    /// tree is empty in a unit test, so the render tests compare pictures.
    @MainActor
    func picture(_ view: some View) -> Data? {
        let host = UIHostingController(rootView: view)
        let window = UIWindow(frame: CGRect(x: 0, y: 0, width: 393, height: 852))
        window.rootViewController = host
        window.makeKeyAndVisible()
        host.view.layoutIfNeeded()
        return UIGraphicsImageRenderer(bounds: host.view.bounds).image { _ in
            host.view.drawHierarchy(in: host.view.bounds, afterScreenUpdates: true)
        }.pngData()
    }
}
