import SwiftUI

/// The eight Category glyphs of `Icon.tsx`, drawn from the shapes in `glyphs.tsx` on its 16-unit grid (PRD D20): each
/// in its Category hue, or in `color` where the glyph sits on ink, a selected tone Chip.
struct Icon: View {
    /// The eight Categories in PRD D20's order, each with its hue (`CATEGORY_TONE`) and its label.
    enum Name: CaseIterable {
        case snacks, games, toys, rides, clothes, bills, gifts, other

        var label: String {
            switch self {
            case .snacks: "Snacks"
            case .games: "Games"
            case .toys: "Toys"
            case .rides: "Rides"
            case .clothes: "Clothes"
            case .bills: "Bills"
            case .gifts: "Gifts"
            case .other: "Other"
            }
        }

        /// The glyph's own fill, `color.data.category-*`.
        var hue: Color {
            switch self {
            case .snacks: Tokens.Data.categorySnacks
            case .games: Tokens.Data.categoryGames
            case .toys: Tokens.Data.categoryToys
            case .rides: Tokens.Data.categoryRides
            case .clothes: Tokens.Data.categoryClothes
            case .bills: Tokens.Data.categoryBills
            case .gifts: Tokens.Data.categoryGifts
            case .other: Tokens.Data.categoryOther
            }
        }

        /// The hue's `*.100` tint a tone Chip sits on.
        var well: Color {
            switch self {
            case .snacks: Tokens.Surface.wellCoral
            case .games: Tokens.Surface.wellViolet
            case .toys: Tokens.Surface.wellYellow
            case .rides: Tokens.Surface.wellSky
            case .clothes: Tokens.Surface.wellGreen
            case .bills: Tokens.Surface.wellInk
            case .gifts: Tokens.Surface.wellGold
            case .other: Tokens.Surface.wellRust
            }
        }
    }

    let name: Name
    var color: Color?
    var size = Tokens.Size.iconGlyph

    var body: some View {
        Canvas { context, canvas in
            context.scaleBy(x: canvas.width / 16, y: canvas.height / 16)
            for part in Self.parts(name) {
                context.fill(part, with: .color(color ?? name.hue))
            }
            if name == .gifts {
                context.fill(Path(CGRect(x: 6.9, y: 4.4, width: 2.2, height: 10.1)), with: .color(Tokens.Palette.white))
            }
        }
        .frame(width: size, height: size)
        .accessibilityHidden(true)
    }

    /// Each glyph as the separate shapes its SVG draws, filled one by one so overlaps never cancel.
    static func parts(_ name: Name) -> [Path] {
        switch name {
        case .snacks:
            [circle(8, 8, 4.2), polygon([(1, 4), (4.4, 8), (1, 12)]), polygon([(15, 4), (11.6, 8), (15, 12)])]
        case .games:
            [circle(8, 3.4, 2.6), Path(CGRect(x: 7, y: 5.4, width: 2, height: 5.2)), dome]
        case .toys:
            [circle(3.9, 4.6, 2.4), circle(12.1, 4.6, 2.4), circle(8, 9.6, 5)]
        case .rides:
            [
                roof,
                Path(roundedRect: CGRect(x: 1, y: 6.3, width: 14, height: 4.6), cornerRadius: 1.8),
                circle(4.4, 12.5, 1.9),
                circle(11.6, 12.5, 1.9),
            ]
        case .clothes:
            [polygon([(6, 2), (10, 2), (14, 4.3), (12.3, 7.4), (11, 6.65), (11, 14), (5, 14), (5, 6.65), (3.7, 7.4),
                      (2, 4.3)])]
        case .bills:
            [polygon([(3, 1.5), (13, 1.5), (13, 12.5), (11.33, 11.4), (9.67, 12.5), (8, 11.4), (6.33, 12.5),
                      (4.67, 11.4), (3, 12.5)])]
        case .gifts:
            [
                circle(5.6, 3, 2.1),
                circle(10.4, 3, 2.1),
                Path(roundedRect: CGRect(x: 1.4, y: 4.4, width: 13.2, height: 3.1), cornerRadius: 1),
                Path(roundedRect: CGRect(x: 2.4, y: 7.5, width: 11.2, height: 7), cornerRadius: 1),
            ]
        case .other:
            [circle(3, 8, 1.8), circle(8, 8, 1.8), circle(13, 8, 1.8)]
        }
    }

    private static func circle(_ x: CGFloat, _ y: CGFloat, _ r: CGFloat) -> Path {
        Path(ellipseIn: CGRect(x: x - r, y: y - r, width: 2 * r, height: 2 * r))
    }

    private static func polygon(_ points: [(CGFloat, CGFloat)]) -> Path {
        Path { path in
            path.addLines(points.map { CGPoint(x: $0.0, y: $0.1) })
            path.closeSubpath()
        }
    }

    /// Games' base, `A6 3.2 0 0 1 14 13.5`: the top half of an ellipse on (2, 13.5) to (14, 13.5), in two quarters.
    private static var dome: Path {
        let k: CGFloat = 0.5523
        return Path { path in
            path.move(to: CGPoint(x: 2, y: 13.5))
            path.addCurve(
                to: CGPoint(x: 8, y: 10.3),
                control1: CGPoint(x: 2, y: 13.5 - 3.2 * k),
                control2: CGPoint(x: 8 - 6 * k, y: 10.3)
            )
            path.addCurve(
                to: CGPoint(x: 14, y: 13.5),
                control1: CGPoint(x: 8 + 6 * k, y: 10.3),
                control2: CGPoint(x: 14, y: 13.5 - 3.2 * k)
            )
            path.closeSubpath()
        }
    }

    /// Rides' cabin: the slanted roof with the two 1.6 corner arcs at its top.
    private static var roof: Path {
        Path { path in
            path.move(to: CGPoint(x: 2.6, y: 6.6))
            path.addLine(to: CGPoint(x: 4.2, y: 3.9))
            path.addQuadCurve(to: CGPoint(x: 5.6, y: 3.1), control: CGPoint(x: 4.6, y: 3.1))
            path.addLine(to: CGPoint(x: 10.4, y: 3.1))
            path.addQuadCurve(to: CGPoint(x: 11.8, y: 3.9), control: CGPoint(x: 11.4, y: 3.1))
            path.addLine(to: CGPoint(x: 13.4, y: 6.6))
            path.closeSubpath()
        }
    }
}
