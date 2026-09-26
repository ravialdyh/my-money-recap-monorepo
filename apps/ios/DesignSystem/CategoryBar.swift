import SwiftUI

/// `CategoryBar` in Bars.tsx: the Category's name and percentage over a 16pt bar in its hue, read by VoiceOver as
/// "{Category}, {n} percent, {amount}" (PRD §11). `shown` reveals the fill (`Reveal`).
struct CategoryBar: View {
    let category: Icon.Name
    let percent: Int
    let amount: String
    var shown = true

    var body: some View {
        VStack(alignment: .leading, spacing: 5) {
            HStack {
                Text(category.label).foregroundStyle(Tokens.Text.primary)
                Spacer(minLength: Tokens.Space._2)
                Text("\(percent)%").foregroundStyle(Tokens.Text.secondary)
            }
            .textStyle(Tokens.TypeRamp.labelS)
            GeometryReader { proxy in
                Capsule()
                    .fill(category.hue)
                    .reveal(shown, width: proxy.size.width * min(max(Double(percent) / 100, 0), 1))
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .background(category.well, in: Capsule())
            }
            .frame(height: Tokens.Size.trackCategory)
        }
        .accessibilityElement(children: .ignore)
        .accessibilityLabel(spoken)
    }

    var spoken: String {
        "\(category.label), \(percent) percent, \(amount)"
    }
}
