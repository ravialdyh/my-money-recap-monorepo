import SwiftUI

/// The Recap pager (spec #151, PRD D46): full screen over Home, its pages swiped left and right, the close control
/// laid over every page top-left. Before the answer it draws the cover's ground and illustration only (§6.19).
struct RecapPagerView: View {
    @Bindable var model: RecapViewModel
    @Environment(\.dismiss) private var dismiss
    @Environment(\.accessibilityReduceMotion) private var reduceMotion

    var body: some View {
        TabView(selection: $model.page) {
            ForEach(Array(0 ..< max(model.pages.count, 1)), id: \.self) { index in
                RecapCoverPage(model: model).tag(index)
            }
        }
        .tabViewStyle(.page(indexDisplayMode: .never))
        .animation(reveal, value: model.answer != nil)
        .overlay(alignment: .topLeading) {
            CloseControl { dismiss() }
                .padding(.leading, 10)
                .padding(.top, 1)
        }
        .background(Tokens.Ground.recap.ignoresSafeArea())
        .task { await model.load() }
    }

    /// `motion.reveal.number` for the numbers arriving; with Reduce Motion, the 150 ms fade.
    private var reveal: Animation {
        reduceMotion
            ? .easeInOut(duration: Tokens.Motion.Duration.reduced)
            : .timingCurve(Tokens.Motion.Easing.standard, duration: Tokens.Motion.Duration.revealNumber)
            .delay(Tokens.Motion.Duration.revealNumberDelay)
    }
}
