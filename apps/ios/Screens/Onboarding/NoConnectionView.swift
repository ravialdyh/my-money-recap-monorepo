import SwiftUI

struct NoConnectionView: View {
    static let copy = "No connection. We need one to save your answers."

    let retrying: Bool
    let retry: () -> Void

    var body: some View {
        Screen(scrolls: true) {
            Spacer(minLength: 0)
            Card {
                InlineError(text: Self.copy)
            }
            Spacer(minLength: 0)
            MMRButton(title: "Try again", loading: retrying, action: retry)
                .disabled(retrying)
        }
    }
}
