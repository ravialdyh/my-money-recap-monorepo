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

/// Nunito body line under a headline, muted on the ground (`SubCopy` in Headline.tsx).
struct SubCopy: View {
    let text: String

    var body: some View {
        Text(text)
            .textStyle(Tokens.TypeRamp.bodyM)
            .foregroundStyle(Tokens.Text.onGroundMuted)
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
