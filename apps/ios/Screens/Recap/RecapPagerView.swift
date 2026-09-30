import SwiftUI

/// The Recap pager (spec #151, PRD D46): full screen over Home, its pages swiped left and right, the close control
/// laid over every page top-left, a target card's D4 over it (#172). Before the answer: cover ground, illustration
/// (§6.19).
struct RecapPagerView: View {
    @Bindable var model: RecapViewModel
    @Environment(\.dismiss) private var dismiss
    @Environment(\.accessibilityReduceMotion) private var reduceMotion

    var body: some View {
        TabView(selection: $model.page) {
            ForEach(Array(0 ..< max(model.pages.count, 1)), id: \.self) { index in
                page(index).tag(index)
            }
        }
        .tabViewStyle(.page(indexDisplayMode: .never))
        .animation(reveal, value: model.answer != nil)
        .overlay(alignment: .topLeading) {
            CloseControl(onLight: ground == Tokens.Ground.earning) { dismiss() }
                .padding(.leading, RecapLayout.closeInset)
                .padding(.top, RecapLayout.closeTop)
        }
        .background(ground.ignoresSafeArea())
        .fullScreenCover(item: $model.targetEdit) { TargetEditView(model: $0) { model.targetEdit = nil } }
        .task { await model.load() }
    }

    /// Each page drawn by its case; the cover before the answer, when there is none.
    @ViewBuilder private func page(_ index: Int) -> some View {
        switch model.pages.indices.contains(index) ? model.pages[index] : .cover {
        case .cover, .partialCover: RecapCoverPage(model: model)
        case .kept, .streak: RecapKeptPage(model: model, active: model.page == index)
        case .goal, .partialGoal: RecapGoalPage(model: model, active: model.page == index)
        case .archetype, .closing: RecapArchetypePage(model: model, active: model.page == index)
        }
    }

    /// The ground behind the safe areas, the page on screen's.
    private var ground: Color {
        switch model.pages.indices.contains(model.page) ? model.pages[model.page] : .cover {
        case .cover, .partialCover: Tokens.Ground.recap
        case .kept, .streak: Tokens.Ground.earning
        case .goal, .partialGoal: Tokens.Ground.goal
        case .archetype, .closing: Tokens.Ground.personality
        }
    }

    /// `motion.reveal.number` for the numbers arriving; with Reduce Motion, the 150 ms fade.
    private var reveal: Animation {
        reduceMotion
            ? .easeInOut(duration: Tokens.Motion.Duration.reduced)
            : .timingCurve(Tokens.Motion.Easing.standard, duration: Tokens.Motion.Duration.revealNumber)
            .delay(Tokens.Motion.Duration.revealNumberDelay)
    }
}
