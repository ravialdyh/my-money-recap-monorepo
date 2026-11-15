import SwiftUI

extension TextStyle {
    var font: Font {
        Font.custom(family, size: size).weight(weight)
    }

    var naturalLineHeight: CGFloat {
        UIFont(descriptor: UIFontDescriptor(fontAttributes: [.family: family]), size: size).lineHeight
    }
}

extension View {
    func textStyle(_ style: TextStyle) -> some View {
        font(style.font)
            .lineSpacing(max(0, style.size * style.lineHeight - style.naturalLineHeight))
            .kerning(style.size * style.letterSpacing)
            .textCase(style.uppercase ? .uppercase : nil)
    }
}
