import SwiftUI

/// A single round Avatar badge for the Home greeting (`Avatar` in AvatarGrid.tsx); an unknown key shows the violet well alone.
struct Avatar: View {
    let key: String
    private let size: CGFloat = 56

    var body: some View {
        Circle()
            .fill(Tokens.Surface.wellViolet)
            .overlay(Image("ill-avatar-\(key)").resizable().scaledToFill())
            .frame(width: size, height: size)
            .clipShape(Circle())
            .accessibilityHidden(true)
    }
}
