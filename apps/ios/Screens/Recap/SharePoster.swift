import CoreText
import SwiftUI

/// The share image (C6, PRD §6.15): five facts and nothing else, so Salary, Money In, kept%, the target, the Streak,
/// the Splurge, the Categories, the Avatar and the email cannot reach it (D28, AC10).
struct SharePoster: Equatable {
    let archetype: Archetype
    /// "September 2026".
    let month: String
    let displayName: String
    /// Saved and Spent as printed, a negative Saved as it is.
    let saved: String
    let spent: String

    /// The kind picks the mascot; the title, "The Squirrel", is ledger's (A2).
    struct Archetype: Equatable {
        let kind: ArchetypePill.Kind
        let title: String
    }

    /// The poster's size in pixels, a phone story's.
    static let size = CGSize(width: 1080, height: 1920)

    /// Where the drawing puts its parts, in pixels, as the C6 and X5 canvases lay them out.
    struct Layout: Equatable {
        let mascot: CGRect
        let lead: CGFloat
        let titleLines: Int
        let card: CGRect
        /// Saved and Spent one per row at 84px: either is too wide for its 348px column at 120px, measured.
        let stacked: Bool
    }

    var layout: Layout {
        let slot = Self.mascotSlot(Self.mascot(archetype.kind))
        let lead = slot.maxY + 24
        let lines = Self.lines(archetype.title, Self.title, width: Self.content, kern: Self.titleKern).count
        let leadLines = Self.lines(leadLine, Self.text(44, .bold), width: Self.content).count
        let nameLines = Self.lines(nameLine, Self.text(44, .heavy), width: Self.content).count
        let cardTop = lead + CGFloat(leadLines) * 52.8 + 4 + CGFloat(lines) * 168 + 28 + CGFloat(nameLines) * 52.8 + 72
        let stacked = [saved, spent].contains { Self.width($0, Self.display(120)) > Self.column }
        let card = CGRect(x: Self.side, y: cardTop, width: Self.content, height: stacked ? 339 : 296)
        return Layout(mascot: slot, lead: lead, titleLines: lines, card: card, stacked: stacked)
    }

    private var leadLine: String {
        "This month, \(displayName), you were"
    }

    private var nameLine: String {
        "\(displayName) · \(month)"
    }

    /// The poster at 1080 by 1920 pixels in the bundled fonts.
    func picture() -> UIImage {
        let format = UIGraphicsImageRendererFormat()
        format.scale = 1
        format.opaque = true
        let layout = layout
        return UIGraphicsImageRenderer(size: Self.size, format: format).image { context in
            UIColor(Tokens.Ground.personality).setFill()
            context.fill(CGRect(origin: .zero, size: Self.size))
            if let mascot = Self.mascot(archetype.kind), let frame = Self.fit(mascot, in: layout.mascot) {
                mascot.draw(in: frame)
            }
            let onGround = UIColor(Tokens.Text.onGround)
            var top = Self.draw(
                leadLine,
                Self.text(44, .bold),
                onGround.withAlphaComponent(0.9),
                top: layout.lead,
                lineHeight: 52.8
            )
            top = Self.draw(archetype.title, Self.title, onGround, top: top + 4, lineHeight: 168, kern: Self.titleKern)
            Self.draw(
                nameLine,
                Self.text(44, .heavy),
                onGround.withAlphaComponent(0.9),
                top: top + 28,
                lineHeight: 52.8
            )
            UIColor(Tokens.Surface.card).setFill()
            UIBezierPath(roundedRect: layout.card, cornerRadius: 56).fill()
            if layout.stacked {
                drawStacked(in: layout.card)
            } else {
                drawSideBySide(in: layout.card)
            }
            Self.draw(
                "MY MONEY RECAP",
                Self.text(34, .heavy),
                onGround.withAlphaComponent(0.85),
                top: Self.size.height - 96 - 40.8,
                lineHeight: 40.8,
                kern: 4.08
            )
        }
    }

    private func drawSideBySide(in card: CGRect) {
        for (index, stat) in stats.enumerated() {
            let left = card.minX + 72 + CGFloat(index) * (Self.column + 48)
            Self.label(stat.label, dot: stat.dot, left: left, top: card.minY + 64)
            Self.draw(
                stat.value,
                Self.display(120),
                UIColor(Tokens.Text.primary),
                top: card.minY + 64 + 36 + 12,
                lineHeight: 120,
                left: left,
                width: Self.column,
                centred: false
            )
        }
    }

    private func drawStacked(in card: CGRect) {
        let inner = card.width - 144
        UIColor(Tokens.Palette.ink100).setFill()
        UIBezierPath(rect: CGRect(x: card.minX + 72, y: card.minY + 56 + 84 + 28, width: inner, height: 3)).fill()
        for (index, stat) in stats.enumerated() {
            let top = card.minY + 56 + CGFloat(index) * (84 + 28 + 3 + 28)
            let labelWidth = Self.label(stat.label, dot: stat.dot, left: card.minX + 72, top: top + 24)
            let room = inner - labelWidth - 32
            let size = min(84, 84 * room / Self.width(stat.value, Self.display(84)))
            let font = Self.display(size)
            let width = Self.width(stat.value, font)
            Self.draw(
                stat.value,
                font,
                UIColor(Tokens.Text.primary),
                top: top + (84 - size) / 2,
                lineHeight: size,
                left: card.maxX - 72 - width,
                width: width,
                centred: false
            )
        }
    }

    private struct Stat {
        let label: String
        let dot: Color
        let value: String
    }

    private var stats: [Stat] {
        [
            Stat(label: "SAVED", dot: Tokens.Data.saved, value: saved),
            Stat(label: "SPENT", dot: Tokens.Data.spent, value: spent),
        ]
    }

    private static let side: CGFloat = 96
    private static let content: CGFloat = size.width - 2 * side
    private static let column: CGFloat = (content - 144 - 48) / 2
    private static let title = display(168)
    private static let titleKern: CGFloat = -1.68

    /// The mascot's 720px slot, as tall as the artwork at that width but never above 620px.
    private static func mascotSlot(_ image: UIImage?) -> CGRect {
        let natural = image.map { 720 * $0.size.height / max($0.size.width, 1) } ?? 620
        return CGRect(x: (size.width - 720) / 2, y: 140, width: 720, height: min(natural, 620))
    }

    /// The artwork scaled to fit `slot` and centred in it, as `object-fit: contain`.
    private static func fit(_ image: UIImage, in slot: CGRect) -> CGRect? {
        guard image.size.width > 0, image.size.height > 0 else { return nil }
        let scale = min(slot.width / image.size.width, slot.height / image.size.height)
        let drawn = CGSize(width: image.size.width * scale, height: image.size.height * scale)
        return CGRect(
            x: slot.midX - drawn.width / 2,
            y: slot.midY - drawn.height / 2,
            width: drawn.width,
            height: drawn.height
        )
    }

    private static func mascot(_ kind: ArchetypePill.Kind) -> UIImage? {
        UIImage(named: RecapArchetypePage.mascot(kind))
    }

    private static func font(_ family: String, _ size: CGFloat, _ weight: UIFont.Weight) -> UIFont {
        let descriptor = UIFontDescriptor(fontAttributes: [.family: family])
            .addingAttributes([.traits: [UIFontDescriptor.TraitKey.weight: weight]])
        return UIFont(descriptor: descriptor, size: size)
    }

    private static func display(_ size: CGFloat) -> UIFont {
        font(Tokens.FontFamily.display, size, .bold)
    }

    private static func text(_ size: CGFloat, _ weight: UIFont.Weight) -> UIFont {
        font(Tokens.FontFamily.text, size, weight)
    }

    /// The dot and the uppercase label in secondary ink; answers the pair's width.
    @discardableResult
    private static func label(_ text: String, dot: Color, left: CGFloat, top: CGFloat) -> CGFloat {
        UIColor(dot).setFill()
        UIBezierPath(ovalIn: CGRect(x: left, y: top + 9, width: 18, height: 18)).fill()
        let font = Self.text(30, .heavy)
        draw(
            text,
            font,
            UIColor(Tokens.Text.secondary),
            top: top,
            lineHeight: 36,
            left: left + 34,
            width: 400,
            centred: false,
            kern: 1.8
        )
        return 34 + width(text, font, kern: 1.8)
    }

    private static func string(
        _ text: String,
        _ font: UIFont,
        _ color: UIColor = .black,
        kern: CGFloat = 0
    ) -> NSAttributedString {
        NSAttributedString(string: text, attributes: [.font: font, .foregroundColor: color, .kern: kern])
    }

    /// The printed width of one line of `text`, as the step-down rule measures it.
    static func width(_ text: String, _ font: UIFont, kern: CGFloat = 0) -> CGFloat {
        CTLineGetTypographicBounds(CTLineCreateWithAttributedString(string(text, font, kern: kern)), nil, nil, nil)
    }

    /// `text` broken into lines no wider than `width`.
    private static func lines(
        _ text: String,
        _ font: UIFont,
        _ color: UIColor = .black,
        width: CGFloat,
        kern: CGFloat = 0
    ) -> [CTLine] {
        let setter = CTFramesetterCreateWithAttributedString(string(text, font, color, kern: kern))
        let path = CGPath(rect: CGRect(x: 0, y: 0, width: width, height: 10000), transform: nil)
        let frame = CTFramesetterCreateFrame(setter, CFRange(location: 0, length: 0), path, nil)
        return (CTFrameGetLines(frame) as? [CTLine]) ?? []
    }

    /// Draws `text` in CSS line boxes of `lineHeight` from `top`, each line's glyphs centred in its box as a browser
    /// sets them; answers the bottom of the last box.
    @discardableResult
    private static func draw(
        _ text: String,
        _ font: UIFont,
        _ color: UIColor,
        top: CGFloat,
        lineHeight: CGFloat,
        left: CGFloat = side,
        width: CGFloat = content,
        centred: Bool = true,
        kern: CGFloat = 0
    ) -> CGFloat {
        guard let context = UIGraphicsGetCurrentContext() else { return top }
        let broken = lines(text, font, color, width: width, kern: kern)
        for (index, line) in broken.enumerated() {
            var ascent: CGFloat = 0
            var descent: CGFloat = 0
            let lineWidth = CTLineGetTypographicBounds(line, &ascent, &descent, nil)
            let trailing = CTLineGetTrailingWhitespaceWidth(line)
            let x = centred ? left + (width - lineWidth + trailing) / 2 : left
            let baseline = top + CGFloat(index) * lineHeight + (lineHeight - ascent - descent) / 2 + ascent
            context.saveGState()
            context.textMatrix = .identity
            context.translateBy(x: x, y: baseline)
            context.scaleBy(x: 1, y: -1)
            CTLineDraw(line, context)
            context.restoreGState()
        }
        return top + CGFloat(broken.count) * lineHeight
    }
}

/// The poster drawn for one tap on Share, which the system share sheet presents while it is set.
struct SharedPicture: Identifiable {
    let id = UUID()
    let image: UIImage
}

/// The system share sheet (C5, IOS I8) with the poster; `done` runs when it closes, whatever the Member chose.
struct ShareSheet: UIViewControllerRepresentable {
    let image: UIImage
    let done: () -> Void

    func makeUIViewController(context _: Context) -> UIActivityViewController {
        let sheet = UIActivityViewController(activityItems: [image], applicationActivities: nil)
        sheet.completionWithItemsHandler = { _, _, _, _ in done() }
        return sheet
    }

    func updateUIViewController(_: UIActivityViewController, context _: Context) {}
}
