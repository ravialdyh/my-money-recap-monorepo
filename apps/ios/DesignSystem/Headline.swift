import SwiftUI

/// Fredoka headline; `l` is the sign-in size (PRD §6.1).
struct Headline: View {
    enum Size { case l, m }
    let text: String
    var size: Size = .l

    var body: some View {
        Text(text)
            .textStyle(size == .l ? Tokens.TypeRamp.headlineL : Tokens.TypeRamp.headlineM)
            .fixedSize(horizontal: false, vertical: true)
    }
}
