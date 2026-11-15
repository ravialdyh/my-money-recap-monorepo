import SwiftUI

struct Sheet<Content: View>: View {
    let title: String
    let back: () -> Void
    @ViewBuilder let content: () -> Content

    var body: some View {
        Screen(scrolls: true) {
            VStack(alignment: .leading, spacing: Tokens.Space._6) {
                VStack(alignment: .leading, spacing: Tokens.Space._1) {
                    Button(action: back) {
                        Image(systemName: "chevron.left")
                            .font(.system(size: 20, weight: .bold))
                            .frame(width: Tokens.Size.touchTargetMin, height: Tokens.Size.touchTargetMin)
                            .contentShape(Rectangle())
                    }
                    .buttonStyle(PressableStyle())
                    .padding(.leading, -14)
                    .accessibilityLabel("Back")
                    Headline(text: title).accessibilityAddTraits(.isHeader)
                }
                content()
            }
            .foregroundStyle(Tokens.Text.primary)
            .padding(.top, Tokens.Space._8)
            .padding(.horizontal, Tokens.Space._8)
            .padding(.bottom, Tokens.Space._9)
            .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .top)
            .background(
                Tokens.Surface.sheet,
                in: RoundedRectangle(cornerRadius: Tokens.Radius.card, style: .continuous)
            )
        }
    }
}
