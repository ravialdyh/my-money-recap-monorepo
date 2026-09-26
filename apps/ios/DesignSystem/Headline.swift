import SwiftUI

/// Fredoka headline, `m` by default as in Headline.tsx; `l` is the sign-in size (PRD §6.1), `compact` the Recap's
/// 36pt size.
struct Headline: View {
    enum Size { case l, m, compact }
    let text: String
    var size: Size = .m

    var body: some View {
        Text(text)
            .textStyle(style)
            .fixedSize(horizontal: false, vertical: true)
    }

    private var style: TextStyle {
        switch size {
        case .l: Tokens.TypeRamp.headlineL
        case .m: Tokens.TypeRamp.headlineM
        case .compact: Tokens.TypeRamp.headlineMCompact
        }
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

/// Small label above a stat (`Eyebrow` in Headline.tsx): Nunito label, secondary ink; `coral` is the uppercase coral
/// tone.
struct Eyebrow: View {
    let text: String
    var coral = false

    var body: some View {
        Text(text)
            .textStyle(coral ? Tokens.TypeRamp.captionUpper : Tokens.TypeRamp.labelM)
            .foregroundStyle(coral ? Tokens.Text.accentCoral : Tokens.Text.secondary)
    }
}
