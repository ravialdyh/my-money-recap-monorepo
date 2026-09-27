import SwiftUI

/// Hero artwork from the asset catalog; hangs `hang`, `space.overlap` unless set, into the card below it. It fills the
/// free height above that card unless `grows` is false (Recap page 2); `maxHeight` caps a tall one, as
/// Illustration.tsx.
struct Illustration: View {
    let name: String
    let width: CGFloat
    var hang = Tokens.Space.overlap
    var grows = true
    var maxHeight: CGFloat?

    var body: some View {
        Image(name)
            .resizable()
            .scaledToFit()
            .frame(width: maxHeight == nil ? width : nil)
            .frame(maxWidth: width, maxHeight: maxHeight)
            .frame(maxWidth: .infinity, maxHeight: grows ? .infinity : nil, alignment: .bottom)
            .padding(.bottom, -hang)
            .zIndex(1)
            .accessibilityHidden(true)
    }
}
