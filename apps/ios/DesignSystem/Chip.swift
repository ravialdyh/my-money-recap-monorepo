import SwiftUI

/// Uppercase caption pill, about 30pt tall. `tappable` lifts the hit area to 44pt; `asLabel` draws it with no tap,
/// the step label; `tone` is the Category selection chip (D70), its glyph on the hue's tint, ink when selected.
struct Chip: View {
    let text: String
    var tappable = false
    var asLabel = false
    var tone: Icon.Name?
    var selected = false
    var action: () -> Void = {}

    var body: some View {
        if asLabel {
            pill
        } else {
            Button(action: action) {
                pill
                    .frame(minHeight: tappable ? Tokens.Size.touchTargetMin : 0)
                    .contentShape(Rectangle())
            }
            .buttonStyle(PressableStyle())
            .accessibilityAddTraits(selected ? .isSelected : [])
        }
    }

    private var pill: some View {
        HStack(spacing: Tokens.Space._2) {
            if let tone {
                Icon(name: tone, color: selected ? Tokens.Palette.white : nil)
            }
            Text(text).textStyle(Tokens.TypeRamp.captionUpper)
        }
        .padding(.horizontal, Tokens.Space._5)
        .padding(.vertical, 8)
        .background(background, in: Capsule())
        .foregroundStyle(selected ? Tokens.Text.onInverse : Tokens.Text.primary)
    }

    private var background: Color {
        guard let tone else { return Tokens.Surface.chipOnGround }
        return selected ? Tokens.Surface.chipInverse : tone.well
    }
}
