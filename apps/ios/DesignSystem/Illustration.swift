import SwiftUI

/// Hero artwork from the asset catalog; hangs `space.overlap` into the card below it.
struct Illustration: View {
    let name: String
    let width: CGFloat

    var body: some View {
        Image(name)
            .resizable()
            .scaledToFit()
            .frame(width: width)
            .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .bottom)
            .padding(.bottom, -Tokens.Space.overlap)
            .zIndex(1)
            .accessibilityHidden(true)
    }
}
