import SwiftUI

struct Screen<Content: View>: View {
    var ground = Tokens.Ground.recap
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
        .foregroundStyle(ground == Tokens.Ground.earning ? Tokens.Text.onGroundDark : Tokens.Text.onGround)
    }

    private var column: some View {
        VStack(alignment: .leading, spacing: 0, content: content)
            .padding(.horizontal, 24)
            .padding(.bottom, 4)
            .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .top)
    }
}
