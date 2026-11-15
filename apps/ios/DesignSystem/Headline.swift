import SwiftUI

struct Headline: View {
    enum Size { case l, m, compact, display }
    let text: String
    var size: Size = .m

    var body: some View {
        if size == .display || size == .l {
            TightLines(text: text, style: style)
        } else {
            Text(text)
                .textStyle(style)
                .fixedSize(horizontal: false, vertical: true)
        }
    }

    private var style: TextStyle {
        switch size {
        case .l: Tokens.TypeRamp.headlineL
        case .m: Tokens.TypeRamp.headlineM
        case .compact: Tokens.TypeRamp.headlineMCompact
        case .display: Tokens.TypeRamp.displayL
        }
    }
}

struct SubCopy: View {
    let text: String

    var body: some View {
        Text(text)
            .textStyle(Tokens.TypeRamp.bodyM)
            .foregroundStyle(Tokens.Text.onGroundMuted)
            .fixedSize(horizontal: false, vertical: true)
    }
}

struct Eyebrow: View {
    let text: String
    var coral = false

    var body: some View {
        Text(text)
            .textStyle(coral ? Tokens.TypeRamp.captionUpper : Tokens.TypeRamp.labelM)
            .foregroundStyle(coral ? Tokens.Text.accentCoral : Tokens.Text.secondary)
    }
}

private struct TightLines: UIViewRepresentable {
    let text: String
    let style: TextStyle

    func makeUIView(context _: Context) -> UILabel {
        let label = UILabel()
        label.numberOfLines = 0
        label.adjustsFontForContentSizeCategory = true
        label.textColor = UIColor(Tokens.Text.onGround)
        return label
    }

    func updateUIView(_ label: UILabel, context _: Context) {
        let family = UIFontDescriptor(fontAttributes: [.family: style.family])
        let bold = family.addingAttributes([.traits: [UIFontDescriptor.TraitKey.weight: UIFont.Weight.bold]])
        label.font = UIFontMetrics.default.scaledFont(for: UIFont(descriptor: bold, size: style.size))
        let lines = NSMutableParagraphStyle()
        lines.lineHeightMultiple = style.size * style.lineHeight / style.naturalLineHeight
        label.attributedText = NSAttributedString(string: text, attributes: [
            .paragraphStyle: lines,
            .kern: style.size * style.letterSpacing,
        ])
    }

    func sizeThatFits(_ proposal: ProposedViewSize, uiView label: UILabel, context _: Context) -> CGSize? {
        let fitted = label.sizeThatFits(CGSize(width: proposal.width ?? .greatestFiniteMagnitude, height: .infinity))
        return CGSize(width: proposal.width ?? fitted.width, height: fitted.height)
    }
}
