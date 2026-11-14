import Connect
import MoneyRecapContract

/// Every server call Settings and the screens it opens make (IOS I3): one box, so a new Settings call is one field
/// here.
/// Each default answers as with no signal, so a test or a picture names only the calls it drives.
struct SettingsCalls {
    var fetch: SettingsViewModel.Fetch = { throw ConnectError(code: .unavailable, message: nil) }
    var saveProfile: NameAvatarViewModel.Save = { _, _ in .offline }
    var saveSalary: SalaryEditViewModel.Save = unavailable
    var saveTarget: TargetEditViewModel.Save = unavailable
    var previewCurrency: CurrencyEditViewModel.Preview = unavailable
    var changeCurrency: CurrencyEditViewModel.Change = unavailable
    var saveNotifications: NotificationsViewModel.Save = { _, _ in nil }
    var requestExport: ExportViewModel.Send = unavailable
    var requestDeletion: DeleteAccountViewModel.Request = unavailable
    var logOut: () async -> Void = {}

    /// The signed-in BFF client's calls, each retried once after a token refresh (IOS I3); Log out forgets the phone's
    /// push token first.
    static func live(
        session: AuthSession,
        client: MoneyRecap_V1_BffServiceClient,
        push: PushRegistration
    ) -> SettingsCalls {
        SettingsCalls(fetch: {
            try await BffClient.retried(session) { await client.getSettings(request: .init()) }.result.get().settings
        }, saveProfile: { request, headers in
            await SaveResult(BffClient.retried(session) { await client.saveProfile(request: request, headers: headers) }
                .result)
        }, saveSalary: { request, headers in
            await BffClient.retried(session) { await client.saveSalary(request: request, headers: headers) }.result
        }, saveTarget: { request, headers in
            await BffClient.retried(session) { await client.saveTarget(request: request, headers: headers) }.result
        }, previewCurrency: { request in
            await BffClient.retried(session) { await client.previewCurrencyChange(request: request) }.result
        }, changeCurrency: { request, headers in
            await BffClient.retried(session) { await client.changeCurrency(request: request, headers: headers) }.result
        }, saveNotifications: { prefs, headers in
            try? await BffClient.retried(session) {
                await client.saveNotificationPrefs(request: .with { $0.prefs = prefs }, headers: headers)
            }.result.get().prefs
        }, requestExport: { headers in
            await BffClient.retried(session) { await client.requestExport(request: .init(), headers: headers) }.result
        }, requestDeletion: { headers in
            await BffClient.retried(session) { await client.requestDeletion(request: .init(), headers: headers) }.result
        }, logOut: {
            await session.logOut(forget: push.forget)
        })
    }
}
