import SwiftUI
import XCTest

extension XCTestCase {
    @MainActor
    func picture(_ view: some View, size: CGSize = CGSize(width: 393, height: 852)) -> Data? {
        let host = hosted(view, size: size)
        return UIGraphicsImageRenderer(bounds: host.view.bounds).image { context in
            host.view.layer.render(in: context.cgContext)
        }.pngData()
    }

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
