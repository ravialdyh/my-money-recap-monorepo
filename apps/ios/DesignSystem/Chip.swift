import SwiftUI

/// Uppercase caption pill in screen headers. Drawn about 30pt tall; `tappable` lifts the hit area to 44pt
/// without changing the pill (README "Tap targets are 44pt").
struct Chip: View {
    enum Variant { case onGround, inverse, highlight }
    let text: String
    var variant: Variant = .onGround
    var tappable = false
    var action: () -> Void = {}

    var body: some View {
        Button(action: action) {
            Text(text)
                .textStyle(Tokens.TypeRamp.captionUpper)
                .padding(.horizontal, 14)
                .padding(.vertical, 8)
                .background(background, in: Capsule())
                .foregroundStyle(variant == .inverse ? Tokens.Text.onInverse : Tokens.Text.primary)
                .frame(minHeight: tappable ? 44 : 0)
                .contentShape(Rectangle())
        }
        .buttonStyle(PressableStyle())
    }

    private var background: Color {
        switch variant {
        case .onGround: Tokens.Surface.chipOnGround
        case .inverse: Tokens.Surface.chipInverse
        case .highlight: Tokens.Surface.chipHighlight
        }
    }
}

/// The header row: chips 8pt apart.
struct ChipRow<Content: View>: View {
    @ViewBuilder let content: () -> Content

    var body: some View {
        HStack(spacing: 8, content: content)
    }
}
