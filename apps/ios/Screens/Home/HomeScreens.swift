import Connect
import Foundation
import MoneyRecapContract
import SwiftData

/// Builds the screens Home opens, so Home shows only Home. A screen using its whole box takes the box
/// (`SettingsViewModel(calls:)`); one using one piece takes only that piece (the Spend list `spendCalls.list`, the
/// Recap `recapCalls.fetch`).
@MainActor
struct HomeScreens {
    /// A write's Home, which replaces the one on screen.
    typealias HomeSink = (MoneyRecap_V1_HomeView) -> Void

    let context: ModelContext
    /// Dates and money print in the device locale (PRD §11); the tests pass a fixed one so the runner's locale cannot
    /// change the copy.
    var locale: Locale = .autoupdatingCurrent
    var connection: Connection = .shared
    var push: PushRegistration?
    var spendCalls = SpendCalls()
    var recapCalls = RecapCalls()
    var settingsCalls = SettingsCalls()

    /// Every box over the signed-in BFF client.
    static func live(
        session: AuthSession,
        client: MoneyRecap_V1_BffServiceClient,
        context: ModelContext
    ) -> HomeScreens {
        let push = PushRegistration.live(session: session, client: client)
        return HomeScreens(
            context: context,
            push: push,
            spendCalls: .live(session: session, client: client),
            recapCalls: .live(session: session, client: client),
            settingsCalls: .live(session: session, client: client, push: push)
        )
    }

    /// The Log a spend sheet in Home's Currency, its decimals and ledger's date range.
    func logSpend(for home: MoneyRecap_V1_HomeView) -> LogSpendViewModel {
        LogSpendViewModel(
            currency: home.currencyCode, digits: Int(home.currencyDigits), earliest: home.earliestSpendDate,
            send: spendCalls.log
        )
    }

    /// B7 Edit spend for one of the list's Spends, over Home's Currency, decimals and date range.
    func editSpend(_ spend: MoneyRecap_V1_Spend, for home: MoneyRecap_V1_HomeView) -> LogSpendViewModel {
        LogSpendViewModel(
            editing: spend, currency: home.currencyCode, digits: Int(home.currencyDigits),
            earliest: home.earliestSpendDate, locale: locale, connection: connection, edit: spendCalls.edit,
            delete: spendCalls.delete
        )
    }

    /// The Spend list of Home's month (PRD §6.10), its switcher open when the month chip opened it (D59).
    func spendList(for home: MoneyRecap_V1_HomeView, switching: Bool) -> SpendListViewModel {
        SpendListViewModel(
            month: home.month, switching: switching, fetch: spendCalls.list, context: context, locale: locale,
            connection: connection
        )
    }

    /// D1 Settings over its stored copy (IOS I4). A saved Target's Home, and a changed Currency's, goes to `onHome`.
    func settings(onHome: @escaping HomeSink) -> SettingsViewModel {
        var calls = settingsCalls
        calls.saveTarget = Self.keeping(calls.saveTarget, onHome: onHome, home: \.home)
        calls.changeCurrency = Self.keeping(calls.changeCurrency, onHome: onHome, home: \.home)
        return SettingsViewModel(calls: calls, push: push, context: context, locale: locale, connection: connection)
    }

    /// D4 over a Recap: Settings read afresh, or its stored copy when that fails; nil with neither.
    func targetEdit(onHome: @escaping HomeSink) async -> TargetEditViewModel? {
        let settings = settings(onHome: onHome)
        await settings.refresh()
        settings.openTarget()
        return settings.targetEdit
    }

    /// One month's Recap pager: a past-Recap row's month, last month, or Home's own month for the preview (D47).
    func recap(_ month: String, onHome: @escaping HomeSink) -> RecapViewModel {
        RecapViewModel(
            month: month, fetch: recapCalls.fetch, context: context, locale: locale,
            openTarget: { await targetEdit(onHome: onHome) }
        )
    }

    private static func keeping<Request, Answer>(
        _ call: @escaping (Request, Headers) async -> Result<Answer, ConnectError>,
        onHome: @escaping HomeSink,
        home: KeyPath<Answer, MoneyRecap_V1_HomeView>
    ) -> (Request, Headers) async -> Result<Answer, ConnectError> {
        { request, headers in
            let result = await call(request, headers)
            if case let .success(answer) = result {
                await onHome(answer[keyPath: home])
            }
            return result
        }
    }
}
