import SwiftUI

/// `PageDots.tsx` with `onSelect`: 6pt dots, the active one 22pt wide, each a control in its own 44pt-tall slice read
/// as "Page {n} of {count}" (PRD D46, §11).
struct PageDots: View {
    let count: Int
    @Binding var active: Int

    var body: some View {
        HStack(spacing: 0) {
            ForEach(0 ..< count, id: \.self) { index in
                Button { active = index } label: {
                    Capsule()
                        .fill(index == active ? Tokens.Indicator.pageActiveOnGround : Tokens.Indicator
                            .pageInactiveOnGround)
                        .frame(
                            width: index == active ? Tokens.Size.pageDotActiveWidth : Tokens.Size.pageDot,
                            height: Tokens.Size.pageDot
                        )
                        .padding(.horizontal, Tokens.Space._2 / 2)
                        .frame(height: Tokens.Size.touchTargetMin)
                        .contentShape(Rectangle())
                }
                .accessibilityLabel("Page \(index + 1) of \(count)")
                .accessibilityAddTraits(index == active ? .isSelected : [])
            }
        }
        .frame(maxWidth: .infinity)
    }
}
