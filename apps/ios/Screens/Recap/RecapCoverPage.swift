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

/// The canvases' exact numbers that no design token holds (`lib.mjs` `closeBtn`, `RecapCoverScreen`, X4 in `Main.mjs`).
enum RecapLayout {
    /// The first chip row's shift right to clear the close control (D46).
    static let chipShift: CGFloat = 34
    /// The close control's inset from the leading edge (`closeBtn` left: 10px).
    static let closeInset: CGFloat = 10
    /// A Chip's drawn height in the canvases (`auto,30px`).
    static let chipHeight: CGFloat = 30
    /// Centres the 44pt close control on the chip row, which sits `space.3` below the top.
    static let closeTop = Tokens.Space._3 + (chipHeight - Tokens.Size.touchTargetMin) / 2
    /// The cover illustration's width, and its width once the stat chips wrap (X4).
    static let cover: CGFloat = 310
    static let coverNarrow: CGFloat = 260
}

private extension View {
    /// Hidden, and out of VoiceOver's reach, until the answer arrives (§6.19).
    func revealed(_ shown: Bool) -> some View {
        opacity(shown ? 1 : 0).accessibilityHidden(!shown)
    }
}
