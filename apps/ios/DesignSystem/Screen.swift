import SwiftUI

/// One ground per screen, from four (README "Design rules"). Content sits inside the screen insets with white type.
struct Screen<Content: View>: View {
    enum Ground { case recap, earning, goal, personality }
    let ground: Ground
    @ViewBuilder let content: () -> Content

    var body: some View {
        ZStack {
            color.ignoresSafeArea()
            VStack(alignment: .leading, spacing: 0, content: content)
                .padding(.horizontal, 24)
                .padding(.bottom, 4)
                .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .top)
        }
        .foregroundStyle(ground == .earning ? Tokens.Text.onGroundDark : Tokens.Text.onGround)
    }

    private var color: Color {
        switch ground {
        case .recap: Tokens.Ground.recap
        case .earning: Tokens.Ground.earning
        case .goal: Tokens.Ground.goal
        case .personality: Tokens.Ground.personality
        }
    }
}
