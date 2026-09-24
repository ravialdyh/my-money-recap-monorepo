import SwiftUI

/// Tinted circle holding a Category glyph (`IconWell` in DataRow.tsx): 30pt around the 16pt glyph, on its hue's tint.
struct IconWell: View {
    let tone: Icon.Name

    var body: some View {
        Icon(name: tone)
            .frame(width: Tokens.Size.iconWell, height: Tokens.Size.iconWell)
            .background(tone.well, in: Circle())
    }
}
