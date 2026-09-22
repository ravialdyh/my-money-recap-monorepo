import SwiftUI

/// The one big number on a screen: eyebrow plus the Fredoka value at 80pt, scaled down to 40pt at most when the
/// printed string does not fit (PRD §11, D71); IOS I7's 80/56/40 steps arrive with the IDR proof.
struct StatHero: View {
    let eyebrow: String
    /// Already printed by the platform formatter (`Money.format`), so the hero never knows a currency's decimals.
    let value: String

    var body: some View {
        VStack(alignment: .leading, spacing: 2) {
            Eyebrow(text: eyebrow)
            Text(value)
                .textStyle(Tokens.TypeRamp.displayXl)
                .monospacedDigit()
                .lineLimit(1)
                .minimumScaleFactor(40 / 80)
                .foregroundStyle(Tokens.Text.primary)
        }
    }
}
