import SwiftUI

/// White card, `radius.card`, no shadow. `underIllustration` adds the top inset an overlapping illustration needs;
/// `minHeight` holds a filled Home card at its measured height (`size.home-hero-card`, `size.home-target-card`; PRD
/// §6.8, D69).
struct Card<Content: View>: View {
    /// Card.tsx's `accent`: violet with white type, the second card on the yellow ground.
    var accent = false
    /// The coral well, D10's This is permanent.
    var coral = false
    var underIllustration = false
    var minHeight: CGFloat?
    /// Card.tsx's `split`: children in a row with no trailing inset, for an illustration at the edge, grown into the
    /// height left to it.
    var split = false
    @ViewBuilder let content: () -> Content

    var body: some View {
        stack
            .frame(maxWidth: .infinity, maxHeight: split ? .infinity : nil, alignment: .leading)
            .padding([.top, .leading, .bottom], Tokens.Space._7)
            .padding(.trailing, split ? 0 : Tokens.Space._7)
            .padding(.top, underIllustration ? 34 : 0)
            .frame(
                maxWidth: .infinity,
                minHeight: minHeight,
                maxHeight: split ? .infinity : nil,
                alignment: .topLeading
            )
            .background(
                accent ? Tokens.Surface.cardAccent : coral ? Tokens.Surface.wellCoral : Tokens.Surface.card,
                in: RoundedRectangle(cornerRadius: Tokens.Radius.card, style: .continuous)
            )
            .foregroundStyle(accent ? Tokens.Text.onInverse : Tokens.Text.primary)
    }

    @ViewBuilder private var stack: some View {
        if split {
            HStack(alignment: .bottom, spacing: Tokens.Space._2, content: content)
        } else {
            VStack(alignment: .leading, spacing: Tokens.Space._5, content: content)
        }
    }
}

/// Fredoka title inside a card (`CardTitle` in Card.tsx).
struct CardTitle: View {
    let text: String

    var body: some View {
        Text(text).textStyle(Tokens.TypeRamp.titleM).fixedSize(horizontal: false, vertical: true)
    }
}
