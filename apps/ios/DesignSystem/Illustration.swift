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
            .frame(width: width, height: height)
            .frame(maxWidth: .infinity, minHeight: height, maxHeight: grows ? .infinity : nil, alignment: .bottom)
            .padding(.bottom, -hang)
            .zIndex(1)
            .accessibilityHidden(true)
    }

    /// The drawn height, which never shrinks to fit, as an image in a flex column: the artwork's at `width`, or the
    /// cap.
    private var height: CGFloat? {
        guard let size = UIImage(named: name)?.size, size.width > 0 else { return nil }
        return min(maxHeight ?? .infinity, width * size.height / size.width)
    }
}
