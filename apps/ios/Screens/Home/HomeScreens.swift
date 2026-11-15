import Connect
import Foundation
import MoneyRecapContract
import SwiftData

@MainActor
struct HomeScreens {
    typealias HomeSink = (MoneyRecap_V1_HomeView) -> Void

    let context: ModelContext
    var locale: Locale = .autoupdatingCurrent
    var connection: Connection = .shared
    var push: PushRegistration?
    var spendCalls = SpendCalls()
    var recapCalls = RecapCalls()
    var settingsCalls = SettingsCalls()

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

    func logSpend(for home: MoneyRecap_V1_HomeView) -> LogSpendViewModel {
        LogSpendViewModel(
            currency: home.currencyCode, digits: Int(home.currencyDigits), earliest: home.earliestSpendDate,
            send: spendCalls.log
        )
    }

    func editSpend(_ spend: MoneyRecap_V1_Spend, for home: MoneyRecap_V1_HomeView) -> LogSpendViewModel {
        LogSpendViewModel(
            editing: spend, currency: home.currencyCode, digits: Int(home.currencyDigits),
            earliest: home.earliestSpendDate, locale: locale, connection: connection, edit: spendCalls.edit,
            delete: spendCalls.delete
        )
    }

    func spendList(for home: MoneyRecap_V1_HomeView, switching: Bool) -> SpendListViewModel {
        SpendListViewModel(
            month: home.month, switching: switching, fetch: spendCalls.list, context: context, locale: locale,
            connection: connection
        )
    }

    func settings(onHome: @escaping HomeSink) -> SettingsViewModel {
        var calls = settingsCalls
        calls.saveTarget = Self.keeping(calls.saveTarget, onHome: onHome, home: \.home)
        calls.changeCurrency = Self.keeping(calls.changeCurrency, onHome: onHome, home: \.home)
        return SettingsViewModel(calls: calls, push: push, context: context, locale: locale, connection: connection)
    }

    func targetEdit(onHome: @escaping HomeSink) async -> TargetEditViewModel? {
        let settings = settings(onHome: onHome)
        await settings.refresh()
        settings.openTarget()
        return settings.targetEdit
    }

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
