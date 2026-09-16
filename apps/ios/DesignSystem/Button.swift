import SwiftUI

/// 52pt full-width pill. Ink by default; `secondary` is the white pill with the ink ring.
/// `loading` swaps in a spinner (SCREEN-STATES A1 loading); the environment's `isEnabled` gives the 40% disabled look.
struct MMRButton: View {
    enum Variant { case primary, secondary }
    let title: String
    var variant: Variant = .primary
    var loading = false
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            HStack(spacing: 12) {
                if loading {
                    ProgressView().tint(foreground)
                }
                Text(title).textStyle(Tokens.TypeRamp.bodyM).fontWeight(.heavy)
            }
            .frame(maxWidth: .infinity, minHeight: 52)
            .foregroundStyle(foreground)
            .background(background, in: Capsule())
            .overlay(Capsule().strokeBorder(Tokens.Palette.ink100, lineWidth: variant == .secondary ? 2 : 0))
        }
        .buttonStyle(PressableStyle())
    }

    private var foreground: Color { variant == .primary ? Tokens.Text.onInverse : Tokens.Text.primary }
    private var background: Color { variant == .primary ? Tokens.Surface.buttonPrimary : Tokens.Surface.buttonSecondary }
}

/// Pressed = scale 0.97 and brightness −8% for `motion.duration.press`; disabled = opacity 0.4. One rule each.
struct PressableStyle: ButtonStyle {
    @Environment(\.isEnabled) private var isEnabled

    func makeBody(configuration: Configuration) -> some View {
        configuration.label
            .scaleEffect(configuration.isPressed ? Tokens.Motion.pressScale : 1)
            .brightness(configuration.isPressed ? -0.08 : 0)
            .opacity(isEnabled ? 1 : Tokens.State.disabledOpacity)
            .animation(.timingCurve(0, 0, 0.2, 1, duration: Tokens.Motion.Duration.press), value: configuration.isPressed)
    }
}
