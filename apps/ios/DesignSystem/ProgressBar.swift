import SwiftUI

/// The 22pt goal track (`ProgressBar` in Bars.tsx): a pill in the tone's light track, filled to `fraction` in the
/// tone's
/// data colour. The fraction is drawing only, clamped to the track (ARCHITECTURE A2).
struct ProgressBar: View {
    enum Tone {
        case violet, green

        var track: Color {
            self == .violet ? Tokens.Surface.trackViolet : Tokens.Surface.trackGreen
        }

        var fill: Color {
            self == .violet ? Tokens.Data.categoryGames : Tokens.Data.saved
        }
    }

    let fraction: Double
    var tone: Tone = .green

    var body: some View {
        GeometryReader { proxy in
            Capsule()
                .fill(tone.fill)
                .frame(width: proxy.size.width * min(max(fraction, 0), 1))
                .frame(maxWidth: .infinity, alignment: .leading)
                .background(tone.track, in: Capsule())
        }
        .frame(height: Tokens.Size.trackGoal)
        .accessibilityHidden(true)
    }
}
