import SwiftUI

/// `StreakDots` in Streak.tsx: one 12pt dot per day in a 15-column grid, the days of `active` (1-based) lit.
struct StreakDots: View {
    let total: Int
    var active: ClosedRange<Int>?

    var body: some View {
        LazyVGrid(
            columns: Array(repeating: GridItem(.flexible(), spacing: Tokens.Space._2), count: 15),
            spacing: Tokens.Space._2
        ) {
            ForEach(0 ..< total, id: \.self) { index in
                Capsule()
                    .fill(active?.contains(index + 1) == true ? Tokens.Data.streakActive : Tokens.Data.streakInactive)
                    .frame(height: Tokens.Size.streakDot)
            }
        }
        .accessibilityElement(children: .ignore)
        .accessibilityLabel("\(active?.count ?? 0) active days of \(total)")
    }
}
