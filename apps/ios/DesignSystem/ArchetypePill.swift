import SwiftUI

/// `ArchetypePill` in ArchetypePill.tsx: a 24pt pill in its Archetype's colour, white type, at 60% unless `selected`,
/// which is full opacity with a 2pt white ring. Read by VoiceOver as its name; not a control here.
struct ArchetypePill: View {
    /// The four money personalities with the legend's fixed labels (`ARCHETYPE_LABEL` in types.ts).
    enum Kind: CaseIterable {
        case squirrel, hustler, treatYourself, slowBurner

        var label: String {
            switch self {
            case .squirrel: "Squirrel"
            case .hustler: "Hustler"
            case .treatYourself: "Treat-Yourself"
            case .slowBurner: "Slow Burner"
            }
        }

        var color: Color {
            switch self {
            case .squirrel: Tokens.Archetype.squirrel
            case .hustler: Tokens.Archetype.hustler
            case .treatYourself: Tokens.Archetype.treatYourself
            case .slowBurner: Tokens.Archetype.slowBurner
            }
        }
    }

    let kind: Kind
    var selected = false

    var body: some View {
        Text(kind.label)
            .textStyle(Tokens.TypeRamp.captionPill)
            .foregroundStyle(Tokens.Text.onInverse)
            .padding(.horizontal, 10)
            .padding(.vertical, Tokens.Space._2)
            .background(kind.color, in: Capsule())
            .overlay(Capsule().inset(by: -2).strokeBorder(Tokens.State.selectedRing, lineWidth: selected ? 2 : 0))
            .opacity(selected ? 1 : Tokens.State.unselectedOpacity)
            .accessibilityAddTraits(selected ? .isSelected : [])
    }
}

/// `ArchetypeRow`: the four pills in order, `space.2` apart, the one `selected` picked out.
struct ArchetypeRow: View {
    var selected: ArchetypePill.Kind?

    var body: some View {
        HStack(spacing: Tokens.Space._2) {
            ForEach(ArchetypePill.Kind.allCases, id: \.self) { ArchetypePill(kind: $0, selected: $0 == selected) }
        }
    }
}
