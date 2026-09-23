import SwiftUI

extension TextStyle {
    /// Fredoka and Nunito by family name (Tokens.FontFamily), both bundled in Resources/Fonts (#106).
    var font: Font {
        Font.custom(family, size: size).weight(weight)
    }
}

extension View {
    /// Applies one `Tokens.TypeRamp` entry: family, weight, size, line height, tracking and case.
    func textStyle(_ style: TextStyle) -> some View {
        font(style.font)
            .lineSpacing(max(0, style.size * (style.lineHeight - 1)))
            .kerning(style.size * style.letterSpacing)
            .textCase(style.uppercase ? .uppercase : nil)
    }
}
