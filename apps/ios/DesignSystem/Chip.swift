import SwiftUI

/// Uppercase caption pill in screen headers, on the ground (the `inverse` and `highlight` variants of Chip.tsx
/// arrive with the screens that use them). Drawn about 30pt tall; `tappable` lifts the hit area to 44pt
/// without changing the pill (README "Tap targets are 44pt").
struct Chip: View {
    let text: String
    var tappable = false
    var action: () -> Void = {}

    var body: some View {
        Button(action: action) {
            Text(text)
                .textStyle(Tokens.TypeRamp.captionUpper)
                .padding(.horizontal, 14)
                .padding(.vertical, 8)
                .background(Tokens.Surface.chipOnGround, in: Capsule())
                .foregroundStyle(Tokens.Text.primary)
                .frame(minHeight: tappable ? 44 : 0)
                .contentShape(Rectangle())
        }
        .buttonStyle(PressableStyle())
    }
}
