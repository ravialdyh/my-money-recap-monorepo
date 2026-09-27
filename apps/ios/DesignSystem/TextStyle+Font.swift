import SwiftUI

extension TextStyle {
    /// Fredoka and Nunito by family name (Tokens.FontFamily), both bundled in Resources/Fonts (#106).
    var font: Font {
        Font.custom(family, size: size).weight(weight)
    }

    /// The family's own line height at `size`, from its font metrics, which the ramp's line height is measured against.
    var naturalLineHeight: CGFloat {
        UIFont(descriptor: UIFontDescriptor(fontAttributes: [.family: family]), size: size).lineHeight
    }
}

extension View {
    /// Applies one `Tokens.TypeRamp` entry: family, weight, size, line height (never below the font's own), tracking
    /// and
    /// case.
    func textStyle(_ style: TextStyle) -> some View {
        font(style.font)
            .lineSpacing(max(0, style.size * style.lineHeight - style.naturalLineHeight))
            .kerning(style.size * style.letterSpacing)
            .textCase(style.uppercase ? .uppercase : nil)
    }
}
