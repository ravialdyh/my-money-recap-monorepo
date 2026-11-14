import SwiftUI

/// D1 Settings (PRD §6.16): four groups of rows with their values. Profile opens D2, Currency D5, Salary and payday
/// D3, Saving target D4, Notifications D8, Export spends D9, Delete account D10, Log out D12. Offline, Log out only.
struct SettingsView: View {
    @State private var model: SettingsViewModel
    @Environment(\.dismiss) private var dismiss
    @State private var loggingOut = false

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
                        ListRow(label: "Currency", value: model.currency) { model.openCurrency() }
                    }
                    group("Money") {
                        ListRow(label: "Salary and payday", value: model.salary) { model.openSalary() }
                        ListRow(label: "Saving target", value: model.target) { model.openTarget() }
                    }
                    group("Alerts") {
                        ListRow(label: "Notifications", value: model.notificationsValue) { model.openNotifications() }
                    }
                }
                .disabled(model.offline)
                group("Your data") {
                    Group {
                        ListRow(label: "Export spends") { model.openExport() }
                        ListRow(label: "Delete account", trailing: .none, destructive: true) {
                            model.openDeleteAccount()
                        }
                    }
                    .disabled(model.offline)
                    ListRow(label: "Log out", trailing: .none) { loggingOut = true }
                        .confirmationDialog("Log out", isPresented: $loggingOut, titleVisibility: .hidden) {
                            Button("Log out", role: .destructive) { Task { await model.logOutNow() } }
                            Button("Cancel", role: .cancel) {}
                        } message: {
                            Text(
                                "Log out of My Money Recap? " +
                                    "You'll need your email or Apple or Google account to get back in."
                            )
                        }
                }
            })
            .toolbar(.hidden, for: .navigationBar)
            .navigationDestination(isPresented: salaryOpen) {
                if let salary = model.salaryEdit {
                    SalaryEditView(model: salary) { model.salaryEdit = nil }
                }
            }
            .navigationDestination(isPresented: targetOpen) {
                if let target = model.targetEdit {
                    TargetEditView(model: target, onSetSalary: model.setSalaryFirst) { model.targetEdit = nil }
                }
            }
            .navigationDestination(isPresented: currencyOpen) {
                if let currency = model.currencyEdit {
                    CurrencyEditView(model: currency)
                }
            }
            .navigationDestination(isPresented: notificationsOpen) {
                if let notifications = model.notifications {
                    NotificationsView(model: notifications)
                }
            }
            .navigationDestination(isPresented: deleteOpen) {
                if let delete = model.deleteAccount {
                    DeleteAccountView(model: delete)
                }
            }
            .navigationDestination(isPresented: exportOpen) {
                if let export = model.export {
                    ExportView(model: export)
                }
            }
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

    /// D8 closes back to D1, which asks again, since the saves answer only the preferences.
    private var notificationsOpen: Binding<Bool> {
        Binding(get: { model.notifications != nil }, set: {
            if !$0 {
                model.notifications = nil
                Task { await model.refresh() }
            }
        })
    }

    private var profileOpen: Binding<Bool> {
        Binding(get: { model.profile != nil }, set: {
            if !$0 {
                model.profile = nil
            }
        })
    }

    private var deleteOpen: Binding<Bool> {
        Binding(get: { model.deleteAccount != nil }, set: {
            if !$0 {
                model.deleteAccount = nil
            }
        })
    }

    private var exportOpen: Binding<Bool> {
        Binding(get: { model.export != nil }, set: {
            if !$0 {
                model.export = nil
            }
        })
    }

    private var currencyOpen: Binding<Bool> {
        Binding(get: { model.currencyEdit != nil }, set: {
            if !$0 {
                model.currencyEdit = nil
            }
        })
    }

    private var targetOpen: Binding<Bool> {
        Binding(get: { model.targetEdit != nil }, set: {
            if !$0 {
                model.targetEdit = nil
            }
        })
    }

    private var salaryOpen: Binding<Bool> {
        Binding(get: { model.salaryEdit != nil }, set: {
            if !$0 {
                model.salaryEdit = nil
            }
        })
    }

    /// An `Eyebrow` over its rows, the canvases' `DataRowList`.
    private func group(_ eyebrow: String, @ViewBuilder rows: () -> some View) -> some View {
        VStack(alignment: .leading, spacing: Tokens.Space._2) {
            Eyebrow(text: eyebrow).accessibilityAddTraits(.isHeader)
            VStack(spacing: 0, content: rows)
        }
    }
}
