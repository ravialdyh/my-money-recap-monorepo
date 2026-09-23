import SwiftUI
import UIKit

/// The one big number on a screen: eyebrow plus the Fredoka value at 80, 56 or 40pt, the largest step at which the
/// printed string fits, measured (PRD §11, D71, IOS I7); past 40pt it scales down rather than clip.
struct StatHero: View {
    let eyebrow: String
    /// Already printed by the platform formatter (`Money.format`), so the hero never knows a currency's decimals.
    let value: String
    @State private var width: CGFloat = 0

    var body: some View {
        VStack(alignment: .leading, spacing: 2) {
            Eyebrow(text: eyebrow)
            Text(value)
                .font(Font.custom(Tokens.FontFamily.display, size: Self.size(for: value, in: width)).weight(.bold))
                .kerning(Self.size(for: value, in: width) * Tokens.TypeRamp.displayXl.letterSpacing)
                .monospacedDigit()
                .lineLimit(1)
                .minimumScaleFactor(0.5)
                .foregroundStyle(Tokens.Text.primary)
                .frame(maxWidth: .infinity, alignment: .leading)
                .onGeometryChange(for: CGFloat.self) { $0.size.width } action: { width = $0 }
        }
    }

    /// The largest hero step at which `value`, in Fredoka bold, fits `width`; 80pt before the row is measured.
    static func size(for value: String, in width: CGFloat) -> CGFloat {
        let steps = [Tokens.TypeRamp.displayXl.size, Tokens.Size.statHeroM, Tokens.Size.statHeroS]
        return steps
            .first {
                width == 0 || (value as NSString).size(withAttributes: [.font: AmountField.font(of: $0)])
                    .width <= width
            }
            ?? Tokens.Size.statHeroS
    }
}
