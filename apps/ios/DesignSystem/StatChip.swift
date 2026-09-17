import SwiftUI

/// Tinted pill with an 8pt status dot: "Money in £2,400", "Saved £640", "Spent £1,760".
struct StatChip: View {
    enum Tone { case earned, saved, spent }
    let tone: Tone
    let text: String

    var body: some View {
        HStack(spacing: 6) {
            Circle().fill(dot).frame(width: 8, height: 8)
            Text(text).textStyle(Tokens.TypeRamp.labelS).lineLimit(1)
        }
        .padding(.vertical, 9)
        .padding(.horizontal, 14)
        .background(background, in: Capsule())
        .foregroundStyle(Tokens.Text.primary)
    }

    private var background: Color {
        switch tone {
        case .earned: Tokens.Data.chipEarned
        case .saved: Tokens.Data.chipSaved
        case .spent: Tokens.Data.chipSpent
        }
    }

    private var dot: Color {
        switch tone {
        case .earned: Tokens.Text.primary
        case .saved: Tokens.Data.saved
        case .spent: Tokens.Data.spent
        }
    }
}
