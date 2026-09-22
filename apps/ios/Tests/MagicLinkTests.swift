import AppAuth
import Connect
import XCTest
@testable import MyMoneyRecap

/// The incoming-link handler's two pure halves: the login token the three sources carry, and the failure
/// code each screen-state copy belongs to (spec #61, SCREEN-STATES A3).
@MainActor
final class MagicLinkTests: XCTestCase {
    func testTheLoginTokenIsReadTheSameWayFromEverySource() throws {
        let token = "lt:f891f3fb-dc63-4e62-92e3-cd0b948b93b8"
        let query = "?lt=lt%3Af891f3fb-dc63-4e62-92e3-cd0b948b93b8"
        let universal = try XCTUnwrap(URL(string: "https://mymoneyrecap.example/magic" + query))
        let scheme = try XCTUnwrap(URL(string: "mymoneyrecap://magic" + query))
        let local = try XCTUnwrap(URL(string: "http://localhost:8090/magic" + query))
        XCTAssertEqual(MagicLink.loginToken(in: universal), token)
        XCTAssertEqual(MagicLink.loginToken(in: scheme), token, "the custom scheme runs the same code")
        XCTAssertEqual(MagicLink.loginToken(in: local), token, "the launch argument hands over this one")
    }

    func testALinkWithoutALoginTokenCarriesNothing() throws {
        XCTAssertNil(try MagicLink.loginToken(in: XCTUnwrap(URL(string: "https://mymoneyrecap.example/magic"))))
        XCTAssertNil(try MagicLink.loginToken(in: XCTUnwrap(URL(string: "mymoneyrecap://auth/callback?code=c"))))
    }

    func testTheRateLimitIsTheOnlyRequestFailureWithItsOwnCopy() {
        XCTAssertNil(MagicLink.failure(requesting: nil), "a sent link is not a failure")
        XCTAssertEqual(
            MagicLink.failure(requesting: ConnectError(code: .resourceExhausted, message: "too many")),
            .rateLimited
        )
        XCTAssertEqual(MagicLink.failure(requesting: ConnectError(code: .unavailable, message: "off")), .generic)
        XCTAssertEqual(MagicLink.failure(requesting: ConnectError(code: .internalError, message: "?")), .generic)
    }

    func testInvalidGrantIsExpiredUntilThisSessionHasRedeemedTheToken() {
        let refused = NSError(
            domain: OIDOAuthTokenErrorDomain,
            code: OIDErrorCodeOAuthToken.invalidGrant.rawValue,
            userInfo: nil
        )
        XCTAssertEqual(
            MagicLink.failure(redeeming: refused, alreadyRedeemed: false),
            .expired,
            "the code cannot tell the two apart, and expired is the default"
        )
        XCTAssertEqual(MagicLink.failure(redeeming: refused, alreadyRedeemed: true), .alreadyUsed)
    }

    func testEveryOtherRedemptionFailureIsTheGenericSignInCopy() {
        let offline = URLError(.notConnectedToInternet)
        let server = NSError(domain: OIDGeneralErrorDomain, code: OIDErrorCode.serverError.rawValue, userInfo: nil)
        XCTAssertEqual(MagicLink.failure(redeeming: offline, alreadyRedeemed: false), .generic)
        XCTAssertEqual(MagicLink.failure(redeeming: server, alreadyRedeemed: true), .generic)
    }

    func testEachCopyIsTheOneSCREENSTATESNames() {
        XCTAssertEqual(MagicLinkFailure.expired.copy, "That link has expired. Links work for 15 minutes.")
        XCTAssertEqual(
            MagicLinkFailure.alreadyUsed.copy,
            "That link has already been used. We can send another."
        )
        XCTAssertEqual(MagicLinkFailure.rateLimited.copy, "That's five links in an hour. Try again later.")
        XCTAssertEqual(MagicLinkFailure.generic.copy, "We couldn't sign you in. Try again.", "A1 error")
        let copies: Set<String> = [
            MagicLinkFailure.expired.copy, MagicLinkFailure.alreadyUsed.copy,
            MagicLinkFailure.rateLimited.copy, MagicLinkFailure.generic.copy,
        ]
        XCTAssertEqual(copies.count, 4, "each failure has its own copy")
    }
}
