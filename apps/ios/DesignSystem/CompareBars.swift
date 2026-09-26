import SwiftUI

/// `CompareBars` in Bars.tsx: 18pt bars, each a label, a track and its amount, read by VoiceOver as "{label},
/// {amount}". The fills reveal (`Reveal`) when `shown` turns true.
struct CompareBars: View {
    struct Row {
        let label: String
        /// The amount as a number, for the drawn length only (ARCHITECTURE A2).
        let value: Double
        let display: String
        let tone: ProgressBar.Tone
    }

    /// `total` reads each row as a share of the whole; `max` scales against the largest row.
    enum Scale { case total, max }

    let rows: [Row]
    var scale = Scale.total
    var shown = true

    var body: some View {
        VStack(spacing: Tokens.Space._3) {
            let fractions = Self.fractions(rows.map(\.value), scale: scale)
            ForEach(rows.indices, id: \.self) { index in
                let row = rows[index], fraction = fractions[index]
                HStack(spacing: Tokens.Space._4) {
                    Text(row.label)
                        .textStyle(Tokens.TypeRamp.labelS)
                        .foregroundStyle(Tokens.Text.secondary)
                        .frame(width: 52, alignment: .leading)
                    GeometryReader { track in
                        Capsule()
                            .fill(row.tone.fill)
                            .reveal(shown, width: track.size.width * fraction)
                            .frame(maxWidth: .infinity, alignment: .leading)
                            .background(row.tone.track, in: Capsule())
                    }
                    .frame(height: Tokens.Size.trackCompare)
                    Text(row.display)
                        .textStyle(Tokens.TypeRamp.labelM)
                        .monospacedDigit()
                        .foregroundStyle(Tokens.Text.primary)
                        .frame(minWidth: 34, alignment: .trailing)
                }
                .accessibilityElement(children: .ignore)
                .accessibilityLabel("\(row.label), \(row.display)")
            }
        }
    }

    /// Each row's drawn length, 0 to 1: its value over the rows' sum or their largest, a negative one drawn empty.
    static func fractions(_ values: [Double], scale: Scale) -> [Double] {
        let basis = scale == .max ? values.max() ?? 0 : values.reduce(0, +)
        return values.map { basis > 0 ? min(max($0 / basis, 0), 1) : 0 }
    }
}
