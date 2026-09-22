import Foundation

/// PRD §6.4's Display Name, counted as `members` counts it: NFC, trimmed, 1 to 30 code points of letters in any
/// script, spaces, hyphens and apostrophes. The field drops the rest as typed and stops at 30 (spec #92).
enum DisplayName {
    static let maximum = 30

    /// What the field keeps of `typed`: its allowed code points in NFC, the first 30.
    static func sanitized(_ typed: String) -> String {
        var kept = String.UnicodeScalarView()
        for scalar in typed.precomposedStringWithCanonicalMapping.unicodeScalars where allowed(scalar) {
            guard kept.count < maximum else { break }
            kept.append(scalar)
        }
        return String(kept)
    }

    /// The name as it is saved, trimmed, or nil while nothing is left.
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
