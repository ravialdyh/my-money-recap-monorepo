import SwiftUI

/// The 22pt goal track (`ProgressBar` in Bars.tsx): a pill in the tone's light track, filled to `fraction` in the
/// tone's data colour, the `piggy` marker riding the fill's end. The fraction is drawing only, clamped (A2).
struct ProgressBar: View {
    enum Tone {
        case violet, green, coral

        var track: Color {
            switch self {
            case .violet: Tokens.Surface.trackViolet
            case .green: Tokens.Surface.trackGreen
            case .coral: Tokens.Surface.trackCoral
            }
        }

        var fill: Color {
            switch self {
            case .violet: Tokens.Data.categoryGames
            case .green: Tokens.Data.saved
            case .coral: Tokens.Data.spent
            }
        }
    }

    let fraction: Double
    var tone: Tone = .green
    var piggy = false
    /// False draws the empty track; turning true reveals the fill (`Reveal`).
    var shown = true

    var body: some View {
        GeometryReader { proxy in
            let end = proxy.size.width * min(max(fraction, 0), 1)
            Capsule()
                .fill(tone.fill)
                .reveal(shown, width: end)
                .frame(maxWidth: .infinity, alignment: .leading)
                .background(tone.track, in: Capsule())
                .overlay(alignment: .bottomLeading) {
                    if piggy {
                        PiggyMarker().offset(x: end - PiggyMarker.size.width / 2, y: -Tokens.Size.trackGoal)
                    }
                }
        }
        .frame(height: Tokens.Size.trackGoal)
        .accessibilityHidden(true)
    }
}

/// The saving-target bar's marker (`PiggyMarker` in glyphs.tsx), drawn from its shapes on the 64 by 30 grid.
struct PiggyMarker: View {
    static let size = CGSize(width: 64, height: 30)

    var body: some View {
        Canvas { context, _ in
            let coral = GraphicsContext.Shading.color(Tokens.Palette.coral500)
            let ink = GraphicsContext.Shading.color(Tokens.Palette.ink900)
            context.fill(
                Path(ellipseIn: CGRect(x: 27.9, y: 0.4, width: 7.2, height: 7.2)),
                with: .color(Tokens.Palette.gold500)
            )
            var tail = Path()
            tail.move(to: CGPoint(x: 16, y: 15))
            tail.addCurve(to: CGPoint(x: 10, y: 18), control1: CGPoint(x: 13, y: 15), control2: CGPoint(x: 13, y: 19))
            context.stroke(tail, with: coral, style: StrokeStyle(lineWidth: 2.4, lineCap: .round))
            var ear = Path()
            ear.addLines([CGPoint(x: 20, y: 8.5), CGPoint(x: 25.4, y: 9.7), CGPoint(x: 24.2, y: 14.7)])
            context.fill(ear, with: coral)
            for x in [22, 35.8] {
                context.fill(
                    Path(roundedRect: CGRect(x: x, y: 23, width: 4.2, height: 6), cornerRadius: 1.7),
                    with: coral
                )
            }
            context.fill(Path(ellipseIn: CGRect(x: 16, y: 7.5, width: 30, height: 19)), with: coral)
            context.fill(Path(ellipseIn: CGRect(x: 40, y: 14.2, width: 10, height: 8.4)), with: coral)
            context.fill(Path(roundedRect: CGRect(x: 26, y: 7.6, width: 11, height: 2.6), cornerRadius: 1.3), with: ink)
            for (x, y, r) in [(37.6, 14.4, 1.5), (43.6, 18.4, 0.9), (46.4, 18.4, 0.9)] {
                context.fill(Path(ellipseIn: CGRect(x: x - r, y: y - r, width: 2 * r, height: 2 * r)), with: ink)
            }
            context.fill(
                Path(ellipseIn: CGRect(x: 34.3, y: 12.5, width: 1.4, height: 1.4)),
                with: .color(Tokens.Palette.white)
            )
        }
        .frame(width: Self.size.width, height: Self.size.height)
        .accessibilityHidden(true)
    }
}

/// A bar's reveal (tokens `motion.reveal.bar`): the fill grows from nothing when `shown` turns true; with Reduce Motion
/// it fades in over 150 ms instead.
struct Reveal: ViewModifier {
    let shown: Bool
    let width: CGFloat
    @Environment(\.accessibilityReduceMotion) private var reduceMotion

    func body(content: Content) -> some View {
        content
            .frame(width: shown || reduceMotion ? width : 0)
            .opacity(shown || !reduceMotion ? 1 : 0)
            .animation(
                reduceMotion ? .easeInOut(duration: Tokens.Motion.Duration.reduced)
                    : .timingCurve(Tokens.Motion.Easing.standard, duration: Tokens.Motion.Duration.revealBar),
                value: shown
            )
    }
}

extension View {
    func reveal(_ shown: Bool, width: CGFloat) -> some View {
        modifier(Reveal(shown: shown, width: width))
    }
}
