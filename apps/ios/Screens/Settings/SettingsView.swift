import SwiftUI

/// D1 Settings (PRD §6.16), the full sheet Home's Settings chip opens: four groups of rows, each with its current
/// value. Profile opens D2; the rows later tickets build are drawn and open nothing yet. Offline, only Log out is live.
struct SettingsView: View {
    @State private var model: SettingsViewModel
    @Environment(\.dismiss) private var dismiss

    init(model: SettingsViewModel) {
        _model = State(initialValue: model)
    }

    var body: some View {
        NavigationStack {
            Sheet(title: "Settings", back: { dismiss() }, content: {
                if model.offline {
                    Banner(tone: .offline, text: "No connection. You can look, but not log.")
                }
                Group {
                    group("You") {
                        ListRow(label: "Profile", value: model.name) { model.openProfile() }
                        ListRow(label: "Currency", value: model.currency)
                    }
                    group("Money") {
                        ListRow(label: "Salary and payday", value: model.salary)
                        ListRow(label: "Saving target", value: model.target)
                    }
                    group("Alerts") {
                        ListRow(label: "Notifications")
                    }
                }
                .disabled(model.offline)
                group("Your data") {
                    Group {
                        ListRow(label: "Export spends")
                        ListRow(label: "Delete account", trailing: .none, destructive: true)
                    }
                    .disabled(model.offline)
                    ListRow(label: "Log out", trailing: .none)
                }
            })
            .toolbar(.hidden, for: .navigationBar)
            .navigationDestination(isPresented: profileOpen) {
                if let profile = model.profile {
                    ProfileView(model: profile) {
                        model.profile = nil
                        Task { await model.refresh() }
                    }
                }
            }
        }
        .task { await model.refresh() }
    }

    private var profileOpen: Binding<Bool> {
        Binding(get: { model.profile != nil }, set: {
            if !$0 {
                model.profile = nil
            }
        })
    }

    /// An `Eyebrow` over its rows, the canvases' `DataRowList`, with no divider under the last.
    private func group(_ eyebrow: String, @ViewBuilder rows: () -> some View) -> some View {
        VStack(alignment: .leading, spacing: Tokens.Space._2) {
            Eyebrow(text: eyebrow).accessibilityAddTraits(.isHeader)
            VStack(spacing: 0, content: rows)
        }
    }
}
