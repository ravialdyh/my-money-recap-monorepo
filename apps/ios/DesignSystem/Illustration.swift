import SwiftUI

/// Hero artwork from the asset catalog; hangs `hang`, `space.overlap` unless set, into the card below it. It fills the
/// free height above that card unless `grows` is false, where the page reads top-down (Recap page 2).
struct Illustration: View {
    let name: String
    let width: CGFloat
    var hang = Tokens.Space.overlap
    var grows = true

    var body: some View {
        Image(name)
            .resizable()
            .scaledToFit()
            .frame(width: width)
            .frame(maxWidth: .infinity, maxHeight: grows ? .infinity : nil, alignment: .bottom)
            .padding(.bottom, -hang)
            .zIndex(1)
            .accessibilityHidden(true)
    }
}
