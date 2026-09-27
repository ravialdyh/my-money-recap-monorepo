import SwiftUI
import XCTest

extension XCTestCase {
    /// `view` settled, every reveal skipped, as it draws in a 393 × 852 window, the size of the iPhone the UI tests run
    /// on, or `size`. SwiftUI's accessibility tree is empty in a unit test, so the render tests compare pictures.
    @MainActor
    func picture(_ view: some View, size: CGSize = CGSize(width: 393, height: 852)) -> Data? {
        let host = hosted(view, size: size)
        return UIGraphicsImageRenderer(bounds: host.view.bounds).image { context in
            host.view.layer.render(in: context.cgContext)
        }.pngData()
    }

    /// How far the vertical scroll view on screen, a pager's neighbouring pages left out, runs past its bounds with its
    /// safe-area insets once `view` is laid out at `size`, to the point: 0 when it fits, above 0 when it scrolls.
    @MainActor
    func overflow(_ view: some View, size: CGSize) -> CGFloat {
        let host = hosted(view, size: size)
        func scrollViews(_ view: UIView) -> [UIScrollView] {
            ((view as? UIScrollView).map { [$0] } ?? []) + view.subviews.flatMap(scrollViews)
        }
        let onScreen = scrollViews(host.view).filter { scroll in
            scroll.contentSize.width <= scroll.bounds.width && abs(scroll.convert(scroll.bounds, to: nil).minX) < 1
        }
        return onScreen.map { scroll in
            let insets = scroll.adjustedContentInset
            return (scroll.contentSize.height + insets.top + insets.bottom - scroll.bounds.height).rounded()
        }.max() ?? 0
    }

    /// The home indicator's inset on an upright iPhone, which a window smaller than the simulator's screen lacks.
    private static var homeIndicator: CGFloat {
        34
    }

    @MainActor
    private func hosted(_ view: some View, size: CGSize) -> UIHostingController<some View> {
        let host = UIHostingController(rootView: view.transaction { $0.animation = nil })
        let window = UIWindow(frame: CGRect(origin: .zero, size: size))
        host.additionalSafeAreaInsets.bottom = size.height > size.width ? Self.homeIndicator : 0
        window.rootViewController = host
        window.makeKeyAndVisible()
        host.view.layoutIfNeeded()
        RunLoop.main.run(until: Date().addingTimeInterval(0.2))
        host.view.layoutIfNeeded()
        return host
    }
}
