import SwiftUI

/// The app's only error surface: a block in the flow inside the card, never an overlay (PRD §6.1).
struct InlineError: View {
    let text: String

    var body: some View {
        HStack(alignment: .top, spacing: 8) {
            Circle()
                .fill(Tokens.Palette.coral500)
                .frame(width: 8, height: 8)
                .padding(.top, 4)
            Text(text).textStyle(Tokens.TypeRamp.labelS)
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(.vertical, 12)
        .padding(.horizontal, 14)
        .background(Tokens.Surface.wellCoral, in: RoundedRectangle(cornerRadius: Tokens.Radius.inset, style: .continuous))
        .foregroundStyle(Tokens.Text.accentCoral)
        .accessibilityAddTraits(.isStaticText)
    }
}
