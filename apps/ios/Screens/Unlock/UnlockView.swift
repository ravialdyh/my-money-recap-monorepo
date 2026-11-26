import SwiftUI

struct UnlockView: View {
    let session: AuthSession

    var body: some View {
        Screen(scrolls: true) {
            Headline(text: "Your month in money.", size: .l)
            SubCopy(text: "Unlock to see your recap.").padding(.top, 10)
            Illustration(name: "ill-cover", width: 260)
            Card(underIllustration: true) {
                MMRButton(title: "Unlock") { Task { await session.unlock() } }
            }
        }
        .task { await session.unlock() }
    }
}
