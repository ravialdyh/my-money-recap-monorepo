import SwiftUI
import XCTest

extension XCTestCase {
    /// `view` as it draws in a 393 × 852 window, the size of the iPhone the UI tests run on, or `size`. SwiftUI's
    /// accessibility tree is empty in a unit test, so the render tests compare pictures.
    @MainActor
    func picture(_ view: some View, size: CGSize = CGSize(width: 393, height: 852)) -> Data? {
        let host = hosted(view, size: size)
        return UIGraphicsImageRenderer(bounds: host.view.bounds).image { _ in
            host.view.drawHierarchy(in: host.view.bounds, afterScreenUpdates: true)
        }.pngData()
    }

    /// How far the tallest vertical scroll view's content runs past its bounds once `view` is laid out at `size`: 0
    /// when everything fits, above 0 when the page scrolls rather than clips.
    @MainActor
    func overflow(_ view: some View, size: CGSize) -> CGFloat {
        let host = hosted(view, size: size)
        func scrollViews(_ view: UIView) -> [UIScrollView] {
            ((view as? UIScrollView).map { [$0] } ?? []) + view.subviews.flatMap(scrollViews)
        }
        return scrollViews(host.view).map { $0.contentSize.height - $0.bounds.height }.max() ?? 0
    }

    @MainActor
    private func hosted(_ view: some View, size: CGSize) -> UIHostingController<some View> {
        let host = UIHostingController(rootView: view)
        let window = UIWindow(frame: CGRect(origin: .zero, size: size))
        window.rootViewController = host
        window.makeKeyAndVisible()
        host.view.layoutIfNeeded()
        RunLoop.main.run(until: Date().addingTimeInterval(0.2))
        host.view.layoutIfNeeded()
        return host
    }
}
