import SwiftUI
import UIKit

struct StatHero: View {
    let eyebrow: String
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
