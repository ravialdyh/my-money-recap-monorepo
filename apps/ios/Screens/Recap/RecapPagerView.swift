import SwiftUI

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

    @ViewBuilder private func page(_ index: Int) -> some View {
        switch model.pages.indices.contains(index) ? model.pages[index] : .cover {
        case .cover, .partialCover: RecapCoverPage(model: model)
        case .kept, .streak: RecapKeptPage(model: model, active: model.page == index)
        case .goal, .partialGoal: RecapGoalPage(model: model, active: model.page == index)
        case .archetype, .closing: RecapArchetypePage(model: model, active: model.page == index)
        }
    }

    private var ground: Color {
        switch model.pages.indices.contains(model.page) ? model.pages[model.page] : .cover {
        case .cover, .partialCover: Tokens.Ground.recap
        case .kept, .streak: Tokens.Ground.earning
        case .goal, .partialGoal: Tokens.Ground.goal
        case .archetype, .closing: Tokens.Ground.personality
        }
    }

    private var reveal: Animation {
        reduceMotion
            ? .easeInOut(duration: Tokens.Motion.Duration.reduced)
            : .timingCurve(Tokens.Motion.Easing.standard, duration: Tokens.Motion.Duration.revealNumber)
            .delay(Tokens.Motion.Duration.revealNumberDelay)
    }
}
