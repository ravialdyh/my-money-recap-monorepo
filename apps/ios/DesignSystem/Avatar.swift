import SwiftUI

struct Avatar: View {
    let key: String
    private let size = Tokens.Size.avatarInline

    var body: some View {
        Circle()
            .fill(Tokens.Surface.wellViolet)
            .overlay(Image("ill-avatar-\(key)").resizable().scaledToFill())
            .frame(width: size, height: size)
            .clipShape(Circle())
            .accessibilityHidden(true)
    }
}
