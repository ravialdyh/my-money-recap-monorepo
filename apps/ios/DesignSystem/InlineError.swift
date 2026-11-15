import SwiftUI

struct InlineError: View {
    let text: String
    var well = Tokens.Surface.wellCoral
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
