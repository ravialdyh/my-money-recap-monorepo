import SwiftUI

struct MMRButton: View {
    enum Variant { case primary, secondary, coral }
    let title: String
    var variant: Variant = .primary
    var icon: String?
    var loading = false
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            HStack(spacing: 12) {
                if loading {
                    ProgressView().tint(foreground)
                } else if let icon {
                    Image(systemName: icon).font(.system(size: 18, weight: .bold))
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

    private var foreground: Color {
        variant == .secondary ? Tokens.Text.primary : Tokens.Text.onInverse
    }

    private var background: Color {
        switch variant {
        case .primary: Tokens.Surface.buttonPrimary
        case .secondary: Tokens.Surface.buttonSecondary
        case .coral: Tokens.Palette.coral500
        }
    }
}

struct PressableStyle: ButtonStyle {
    @Environment(\.isEnabled) private var isEnabled

    func makeBody(configuration: Configuration) -> some View {
        configuration.label
            .scaleEffect(configuration.isPressed ? Tokens.Motion.pressScale : 1)
            .brightness(configuration.isPressed ? Tokens.State.pressedLightnessDelta : 0)
            .opacity(isEnabled ? 1 : Tokens.State.disabledOpacity)
            .animation(
                .timingCurve(Tokens.Motion.Easing.press, duration: Tokens.Motion.Duration.press),
                value: configuration.isPressed
            )
    }
}
