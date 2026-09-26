import SwiftUI

/// Uppercase caption pill, about 30pt tall. `tappable` lifts the hit area to 44pt; `asLabel` draws it with no tap;
/// `tone` is the Category chip (D70); `opens` adds the switcher's chevron (D59); `highlight` is the yellow pill.
struct Chip: View {
    let text: String
    var tappable = false
    var asLabel = false
    var tone: Icon.Name?
    var selected = false
    var opens = false
    var highlight = false
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
            if opens {
                Image(systemName: "chevron.down").font(.system(size: 11, weight: .heavy))
            }
        }
        .padding(.horizontal, Tokens.Space._5)
        .padding(.vertical, 8)
        .background(background, in: Capsule())
        .foregroundStyle(selected ? Tokens.Text.onInverse : Tokens.Text.primary)
    }

    private var background: Color {
        guard let tone else { return highlight ? Tokens.Surface.chipHighlight : Tokens.Surface.chipOnGround }
        return selected ? Tokens.Surface.chipInverse : tone.well
    }
}
