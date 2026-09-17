import SwiftUI

/// A single round Avatar badge for the Home greeting (56pt) and the recap chip (`Avatar` in AvatarGrid.tsx).
/// The six illustrations are not in the asset catalog yet, so an unknown key shows the violet well alone.
struct Avatar: View {
    let key: String
    var size: CGFloat = 56

    var body: some View {
        Circle()
            .fill(Tokens.Surface.wellViolet)
            .overlay(Image("ill-avatar-\(key)").resizable().scaledToFill())
            .frame(width: size, height: size)
            .clipShape(Circle())
            .accessibilityHidden(true)
    }
}
