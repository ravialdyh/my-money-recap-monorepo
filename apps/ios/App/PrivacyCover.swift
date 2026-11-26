import SwiftUI

@MainActor
enum PrivacyCover {
    private static var window: UIWindow?

    static func show() {
        guard window == nil,
              let scene = UIApplication.shared.connectedScenes.compactMap({ $0 as? UIWindowScene }).first
        else { return }
        let cover = UIWindow(windowScene: scene)
        cover.windowLevel = .alert + 1
        cover.rootViewController = UIHostingController(rootView: Screen {
            Headline(text: "Your month in money.", size: .l)
            Illustration(name: "ill-cover", width: 260)
        })
        cover.isHidden = false
        window = cover
    }

    static func hide() {
        window?.isHidden = true
        window = nil
    }
}
