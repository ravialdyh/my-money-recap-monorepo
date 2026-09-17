import SwiftUI

/// One ground per screen, from four (README "Design rules"); every screen so far is the violet `recap` ground, and the
/// other three (with yellow's ink type) arrive with the Recap screens. Content sits inside the screen insets with white
/// type.
struct Screen<Content: View>: View {
    @ViewBuilder let content: () -> Content

    var body: some View {
        ZStack {
            Tokens.Ground.recap.ignoresSafeArea()
            VStack(alignment: .leading, spacing: 0, content: content)
                .padding(.horizontal, 24)
                .padding(.bottom, 4)
                .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .top)
        }
        .foregroundStyle(Tokens.Text.onGround)
    }
}
