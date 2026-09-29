import SwiftUI

/// D8 Notifications (PRD §6.16): the three switches with their helper lines, and `Remind me at` under Daily reminder
/// while it is on, its time the system's own picker in the device locale. Denied, the info Banner and the rows at 40%.
struct NotificationsView: View {
    @State private var model: NotificationsViewModel
    @Environment(\.dismiss) private var dismiss

    init(model: NotificationsViewModel) {
        _model = State(initialValue: model)
    }

    var body: some View {
        Sheet(title: "Notifications", back: { dismiss() }, content: {
            if model.denied {
                Banner(
                    tone: .info,
                    text: "Notifications are off for this app. Turn them on in iOS Settings.",
                    link: ("Open Settings", NotificationsViewModel.openSystemSettings)
                )
            }
            Card {
                VStack(spacing: 0) {
                    row("Recap ready", "09:00 on the 1st, when last month has data.", .recapReady)
                    row("Daily reminder", "Only if you haven't logged a spend that day.", .dailyReminder)
                    if model.isOn(.dailyReminder) {
                        reminderTime
                    }
                    row(
                        "Target warning", "Once a month, the first time you pass your saving line.", .targetWarning,
                        divider: false
                    )
                }
            }
            .opacity(model.denied ? Tokens.State.disabledOpacity : 1)
            .disabled(model.offline)
            if model.failed {
                InlineError(text: SettingsViewModel.saveFailure)
            }
        })
        .toolbar(.hidden, for: .navigationBar)
        .task { await model.readPermission() }
        .onReceive(NotificationCenter.default.publisher(for: UIApplication.didBecomeActiveNotification)) { _ in
            Task { await model.readPermission() }
        }
    }

    private func row(
        _ title: String,
        _ helper: String,
        _ which: NotificationsViewModel.Switch,
        divider: Bool = true
    ) -> some View {
        HStack(spacing: 14) {
            VStack(alignment: .leading, spacing: 2) {
                Text(title).textStyle(Tokens.TypeRamp.labelM)
                Text(helper).textStyle(Tokens.TypeRamp.bodyS).foregroundStyle(Tokens.Text.secondary)
                    .fixedSize(horizontal: false, vertical: true)
            }
            .frame(maxWidth: .infinity, alignment: .leading)
            MMRToggle(label: title, isOn: Binding(
                get: { model.isOn(which) },
                set: { on in Task { await model.flip(which, to: on) } }
            ))
        }
        .padding(.vertical, 14)
        .overlay(alignment: .bottom) { Tokens.Surface.divider.frame(height: divider ? 1 : 0) }
    }

    /// `Remind me at` and the saved time; tapping the time opens the system wheel, and each settled time is saved.
    private var reminderTime: some View {
        HStack(spacing: 12) {
            Text("Remind me at").textStyle(Tokens.TypeRamp.labelM)
            Spacer(minLength: 0)
            DatePicker("Remind me at", selection: Binding(
                get: { model.reminderTime },
                set: { time in Task { await model.pick(time) } }
            ), displayedComponents: .hourAndMinute)
                .labelsHidden()
                .disabled(model.denied)
        }
        .padding(.vertical, Tokens.Space._3)
        .overlay(alignment: .bottom) { Tokens.Surface.divider.frame(height: 1) }
    }
}
