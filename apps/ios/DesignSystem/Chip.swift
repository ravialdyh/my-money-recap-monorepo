import SwiftUI

/// Uppercase caption pill on the ground, about 30pt tall. `tappable` lifts the hit area to 44pt (README "Tap
/// targets are 44pt"); `asLabel` draws the pill with no tap, the onboarding step label (`asLabel` in Chip.tsx).
struct Chip: View {
    let text: String
    var tappable = false
    var asLabel = false
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
        }
    }

    private var pill: some View {
        Text(text)
            .textStyle(Tokens.TypeRamp.captionUpper)
            .padding(.horizontal, Tokens.Space._5)
            .padding(.vertical, 8)
            .background(Tokens.Surface.chipOnGround, in: Capsule())
            .foregroundStyle(Tokens.Text.primary)
    }
}
