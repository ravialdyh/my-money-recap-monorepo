import Foundation

enum DisplayName {
    static let maximum = 30

    static func sanitized(_ typed: String) -> String {
        var kept = String.UnicodeScalarView()
        for scalar in typed.precomposedStringWithCanonicalMapping.unicodeScalars where allowed(scalar) {
            guard kept.count < maximum else { break }
            kept.append(scalar)
        }
        return String(kept)
    }

    static func stored(_ name: String) -> String? {
        let trimmed = sanitized(name).trimmingCharacters(in: .whitespaces)
        return trimmed.isEmpty ? nil : trimmed
    }

    static func allowed(_ scalar: Unicode.Scalar) -> Bool {
        switch scalar.properties.generalCategory {
        case .uppercaseLetter, .lowercaseLetter, .titlecaseLetter, .modifierLetter, .otherLetter,
             .nonspacingMark, .spacingMark, .enclosingMark:
            true
        default:
            " -'’".unicodeScalars.contains(scalar)
        }
    }
}
