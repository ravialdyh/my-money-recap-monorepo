import SwiftUI

/// White card, `radius.card`, no shadow. `underIllustration` adds the top inset an overlapping illustration needs.
struct Card<Content: View>: View {
    var underIllustration = false
    @ViewBuilder let content: () -> Content

    var body: some View {
        VStack(alignment: .leading, spacing: 14, content: content)
            .frame(maxWidth: .infinity)
            .padding(22)
            .padding(.top, underIllustration ? 34 : 0)
            .background(Tokens.Surface.card, in: RoundedRectangle(cornerRadius: Tokens.Radius.card, style: .continuous))
            .foregroundStyle(Tokens.Text.primary)
    }
}
