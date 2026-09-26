import SwiftUI

/// Page 1, C1 or P1 (PRD §6.11, §6.17): the chips, the headline, the cover illustration over the hero card, the hint
/// and the page dots. Before the answer only the illustration shows; the rest fades in. Scrolls held sideways.
struct RecapCoverPage: View {
    @Bindable var model: RecapViewModel
    @State private var chipsWrap = false

    var body: some View {
        Screen(scrolls: true) {
            StatChipRow {
                Chip(text: model.monthChip, asLabel: true)
                Chip(text: model.nameChip, asLabel: true, highlight: true)
            }
            .padding(.leading, 34)
            .padding(.top, 8)
            .revealed(model.answer != nil)
            Headline(text: model.headline, size: .l)
                .padding(.top, 22)
                .revealed(model.answer != nil)
            Illustration(name: "ill-cover", width: chipsWrap ? 260 : 310)
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

private extension View {
    /// Hidden, and out of VoiceOver's reach, until the answer arrives (§6.19).
    func revealed(_ shown: Bool) -> some View {
        opacity(shown ? 1 : 0).accessibilityHidden(!shown)
    }
}
