import SwiftUI

struct ArchetypePill: View {
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

struct ArchetypeRow: View {
    var selected: ArchetypePill.Kind?

    var body: some View {
        HStack(spacing: Tokens.Space._2) {
            ForEach(ArchetypePill.Kind.allCases, id: \.self) { ArchetypePill(kind: $0, selected: $0 == selected) }
        }
    }
}
