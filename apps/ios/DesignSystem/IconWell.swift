import SwiftUI

struct IconWell: View {
    let tone: Icon.Name

    var body: some View {
        Icon(name: tone)
            .frame(width: Tokens.Size.iconWell, height: Tokens.Size.iconWell)
            .background(tone.well, in: Circle())
    }
}
