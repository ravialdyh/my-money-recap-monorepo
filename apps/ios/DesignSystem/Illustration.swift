import SwiftUI

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

    private var height: CGFloat? {
        guard let size = UIImage(named: name)?.size, size.width > 0 else { return nil }
        return min(maxHeight ?? .infinity, width * size.height / size.width)
    }
}
