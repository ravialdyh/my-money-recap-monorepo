import SwiftUI

/// The app's only error surface: a block in the flow inside the card, never an overlay (PRD §6.1).
struct InlineError: View {
    let text: String
    /// The block's fill: coral, or white where the ground is coral itself, the restore sheet (D11b), so it still reads.
    var well = Tokens.Surface.wellCoral
    /// The trailing control (`InlineError.tsx`'s `action`): heavy and underlined, D6's `Try again`; nil removes it.
    var action: (title: String, run: () -> Void)?

    var body: some View {
        HStack(alignment: .top, spacing: 8) {
            Circle()
                .fill(Tokens.Palette.coral500)
                .frame(width: 8, height: 8)
                .padding(.top, 4)
            Text(text).textStyle(Tokens.TypeRamp.labelS)
                .frame(maxWidth: .infinity, alignment: .leading)
                .accessibilityAddTraits(.isStaticText)
            if let action {
                Button(action: action.run) {
                    Text(action.title).textStyle(Tokens.TypeRamp.labelS).fontWeight(.heavy).underline()
                }
                .buttonStyle(PressableStyle())
                .fixedSize()
            }
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(.vertical, 12)
        .padding(.horizontal, Tokens.Space._5)
        .background(
            well,
            in: RoundedRectangle(cornerRadius: Tokens.Radius.inset, style: .continuous)
        )
        .foregroundStyle(Tokens.Text.accentCoral)
    }
}
