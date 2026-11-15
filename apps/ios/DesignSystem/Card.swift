import SwiftUI

struct Card<Content: View>: View {
    var accent = false
    var coral = false
    var underIllustration = false
    var minHeight: CGFloat?
    var split = false
    @ViewBuilder let content: () -> Content

    var body: some View {
        stack
            .frame(maxWidth: .infinity, maxHeight: split ? .infinity : nil, alignment: .leading)
            .padding([.top, .leading, .bottom], Tokens.Space._7)
            .padding(.trailing, split ? 0 : Tokens.Space._7)
            .padding(.top, underIllustration ? 34 : 0)
            .frame(
                maxWidth: .infinity,
                minHeight: minHeight,
                maxHeight: split ? .infinity : nil,
                alignment: .topLeading
            )
            .background(
                accent ? Tokens.Surface.cardAccent : coral ? Tokens.Surface.wellCoral : Tokens.Surface.card,
                in: RoundedRectangle(cornerRadius: Tokens.Radius.card, style: .continuous)
            )
            .foregroundStyle(accent ? Tokens.Text.onInverse : Tokens.Text.primary)
    }

    @ViewBuilder private var stack: some View {
        if split {
            HStack(alignment: .bottom, spacing: Tokens.Space._2, content: content)
        } else {
            VStack(alignment: .leading, spacing: Tokens.Space._5, content: content)
        }
    }
}

struct CardTitle: View {
    let text: String

    var body: some View {
        Text(text).textStyle(Tokens.TypeRamp.titleM).fixedSize(horizontal: false, vertical: true)
    }
}
