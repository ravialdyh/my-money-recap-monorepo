import SwiftUI

/// Signed-in placeholder; the Home ticket replaces it.
struct HomePlaceholderView: View {
    @Environment(AuthSession.self) private var session

    var body: some View {
        Screen(ground: .recap) {
            Headline(text: "You're signed in.")
            SubCopy(text: session.email.map { "Signed in as \($0)." } ?? "Home arrives in the next ticket.").padding(.top, 10)
        }
    }
}
