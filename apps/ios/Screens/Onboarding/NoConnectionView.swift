import SwiftUI

/// Setup's one offline state (SCREEN-STATES A4–A7, spec #92): every answer is saved as given, so it needs a
/// connection. A step shows `copy` in its card and Next tries again; a launch that learns no step shows it here.
struct NoConnectionView: View {
    static let copy = "No connection. We need one to save your answers."

    let retrying: Bool
    let retry: () -> Void

    var body: some View {
        Screen {
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
