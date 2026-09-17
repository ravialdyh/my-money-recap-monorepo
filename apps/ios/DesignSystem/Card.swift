import SwiftUI

/// White card, `radius.card`, no shadow. `underIllustration` adds the top inset an overlapping illustration needs;
/// `minHeight` holds a filled Home card at its measured height (`size.home-hero-card`, `size.home-target-card`; PRD
/// §6.8, D69).
struct Card<Content: View>: View {
    var underIllustration = false
    var minHeight: CGFloat?
    @ViewBuilder let content: () -> Content

    var body: some View {
        VStack(alignment: .leading, spacing: Tokens.Space._5, content: content)
            .frame(maxWidth: .infinity, alignment: .leading)
            .padding(Tokens.Space._7)
            .padding(.top, underIllustration ? 34 : 0)
            .frame(maxWidth: .infinity, minHeight: minHeight, alignment: .topLeading)
            .background(Tokens.Surface.card, in: RoundedRectangle(cornerRadius: Tokens.Radius.card, style: .continuous))
            .foregroundStyle(Tokens.Text.primary)
    }
}

/// Fredoka title inside a card (`CardTitle` in Card.tsx).
struct CardTitle: View {
    let text: String

    var body: some View {
        Text(text).textStyle(Tokens.TypeRamp.titleM).fixedSize(horizontal: false, vertical: true)
    }
}
