import SwiftUI

/// One ground per screen, from four (README "Design rules"): the violet `recap` ground unless `ground` names another.
/// Content sits inside the screen insets with white type.
struct Screen<Content: View>: View {
    var ground = Tokens.Ground.recap
    /// True for a screen with no scroll view of its own: it scrolls when it is taller than the screen, as a phone
    /// held sideways makes it (PRD §11 "Platform"), and fills the screen otherwise, so its spacers still push.
    var scrolls = false
    @ViewBuilder let content: () -> Content

    var body: some View {
        ZStack {
            ground.ignoresSafeArea()
            if scrolls {
                GeometryReader { screen in
                    ScrollView {
                        column.frame(minHeight: screen.size.height)
                    }
                    .scrollBounceBehavior(.basedOnSize)
                }
            } else {
                column
            }
        }
        .foregroundStyle(Tokens.Text.onGround)
    }

    private var column: some View {
        VStack(alignment: .leading, spacing: 0, content: content)
            .padding(.horizontal, 24)
            .padding(.bottom, 4)
            .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .top)
    }
}
