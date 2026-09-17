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

/// Small label above a stat (`Eyebrow` in Headline.tsx): Nunito label, secondary ink.
struct Eyebrow: View {
    let text: String

    var body: some View {
        Text(text).textStyle(Tokens.TypeRamp.labelM).foregroundStyle(Tokens.Text.secondary)
    }
}

/// One sentence under a title inside a card (`HelperText` in Headline.tsx).
struct HelperText: View {
    let text: String

    var body: some View {
        Text(text)
            .textStyle(Tokens.TypeRamp.bodyS)
            .foregroundStyle(Tokens.Text.secondary)
            .fixedSize(horizontal: false, vertical: true)
    }
}
