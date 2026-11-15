import SwiftUI

struct RecapCoverPage: View {
    @Bindable var model: RecapViewModel
    @State private var chipsWrap = false

    var body: some View {
        Screen(scrolls: true) {
            StatChipRow {
                Chip(text: model.monthChip, asLabel: true)
                Chip(text: model.nameChip, asLabel: true, highlight: true)
            }
            .padding(.leading, RecapLayout.chipShift)
            .padding(.top, Tokens.Space._3)
            .revealed(model.answer != nil)
            Headline(text: model.headline, size: .l)
                .padding(.top, Tokens.Space._7)
                .revealed(model.answer != nil)
            Illustration(name: "ill-cover", width: chipsWrap ? RecapLayout.coverNarrow : RecapLayout.cover)
            Card(underIllustration: true) {
                StatHero(eyebrow: model.eyebrow, value: model.hero)
                StatChipRow {
                    ForEach(model.chips, id: \.text) { StatChip(tone: $0.tone, text: $0.text) }
                }
                .onGeometryChange(for: Bool.self) { $0.size.height > Tokens.Size.touchTargetMin } action: {
                    chipsWrap = $0
                }
                Text("Swipe to see where it all went.")
                    .textStyle(Tokens.TypeRamp.bodyS)
                    .foregroundStyle(Tokens.Text.secondary)
            }
            .revealed(model.answer != nil)
            PageDots(count: model.pages.count, active: $model.page)
                .revealed(model.answer != nil)
        }
    }
}

enum RecapLayout {
    static let chipShift: CGFloat = 34
    static let closeInset: CGFloat = 10
    static let chipHeight: CGFloat = 30
    static let closeTop = Tokens.Space._3 + (chipHeight - Tokens.Size.touchTargetMin) / 2
    static let cover: CGFloat = 310
    static let coverNarrow: CGFloat = 260
    static let cardGap: CGFloat = 18
    static let vending: CGFloat = 148
    static let vendingRise: CGFloat = 60
}

private extension View {
    func revealed(_ shown: Bool) -> some View {
        opacity(shown ? 1 : 0).accessibilityHidden(!shown)
    }
}
