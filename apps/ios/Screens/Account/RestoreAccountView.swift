import SwiftUI

/// D11b: the full coral sheet on the violet ground, with no back control, so the Member has to answer it (D60).
/// "Continue with deletion" is only the phone's own sign-out; ending the realm's session comes with Log out.
struct RestoreAccountView: View {
    @State var model: RestoreAccountViewModel
    let onKept: () -> Void
    let onContinue: () -> Void

    var body: some View {
        Screen(scrolls: true) {
            VStack(alignment: .leading, spacing: Tokens.Space._6) {
                Headline(text: "Welcome back")
                    .accessibilityAddTraits(.isHeader)
                Text(model.askedLine)
                    .textStyle(Tokens.TypeRamp.bodyM).fontWeight(.heavy)
                    .fixedSize(horizontal: false, vertical: true)
                Text(model.loseLine)
                    .textStyle(Tokens.TypeRamp.bodyM)
                    .foregroundStyle(Tokens.Text.secondary)
                    .fixedSize(horizontal: false, vertical: true)
                if model.failed {
                    InlineError(text: RestoreAccountViewModel.failedCopy)
                }
                Spacer(minLength: 0)
                VStack(spacing: 10) {
                    MMRButton(title: "Keep my account", loading: model.isKeeping) {
                        Task {
                            if await model.keep() {
                                onKept()
                            }
                        }
                    }
                    MMRButton(title: "Continue with deletion", variant: .secondary, action: onContinue)
                }
                .disabled(model.isKeeping)
            }
            .foregroundStyle(Tokens.Text.primary)
            .padding(.top, Tokens.Space._8)
            .padding(.horizontal, Tokens.Space._8)
            .padding(.bottom, Tokens.Space._9)
            .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .top)
            .background(
                Tokens.Palette.coral100,
                in: RoundedRectangle(cornerRadius: Tokens.Radius.card, style: .continuous)
            )
        }
    }
}
