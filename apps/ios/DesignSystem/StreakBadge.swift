import SwiftUI

struct StreakBadge: View {
    let value: Int
    var unit = "days"

    var body: some View {
        VStack(spacing: 0) {
            Text("\(value)").textStyle(Tokens.TypeRamp.statBadge)
            Text(unit).textStyle(Tokens.TypeRamp.captionUnit)
        }
        .frame(width: Tokens.Size.streakBadge, height: Tokens.Size.streakBadge)
        .background(Tokens.Surface.chipHighlight, in: Circle())
        .foregroundStyle(Tokens.Text.primary)
        .accessibilityElement(children: .ignore)
        .accessibilityLabel("\(value) \(unit)")
    }
}
