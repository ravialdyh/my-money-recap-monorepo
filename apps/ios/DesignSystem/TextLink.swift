import SwiftUI

/// The helper-text link inside a card: violet, underlined, a 44pt tap target (the canvases' "Clear search").
struct TextLink: View {
    let title: String
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            Text(title)
                .textStyle(Tokens.TypeRamp.labelM)
                .underline()
                .foregroundStyle(Tokens.Palette.violet500)
                .padding(.horizontal, 4)
                .frame(minHeight: Tokens.Size.touchTargetMin)
                .contentShape(Rectangle())
        }
        .buttonStyle(PressableStyle())
    }
}
