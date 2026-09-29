# bff

The BFF (`docs/ARCHITECTURE.md` §6): the only thing the phone talks to, owning no database. It
serves `GetHome`, `GetOnboardingState`, `SetTimezone`, `SaveProfile`, `GetSettings`, `SaveSalary`, `SaveTarget`, `CancelDeletion`, `LogSpend`, `EditSpend`, `DeleteSpend`, `ListSpends`, `GetRecap` and `RequestMagicLink`; the rest of
`BffService` is planned. Every RPC but
`RequestMagicLink` passes the bearer check every service shares (`apps/README.md`, "Every
service") before any service is called, and the token is then forwarded unchanged (§4.2).

`RequestMagicLink` (#69) is exempt from that check by procedure name, because it is sent before a
Member has a token, so the handler reads the address from the request and never from the context's
claims. It gets a login token from the realm on A10's terms, as the `bff` service account through
the Phase Two extension's `login-token` endpoint, builds `BFF_MAGIC_LINK_BASE_URL?lt=…` from it, and
hands the address and the link to `notify.EmailMagicLink` (§7.5), exempt on the same rule (#68). The
`Idempotency-Key` header is copied when the caller sends one and is a fresh version 4 UUID when it
does not, because `notify` rejects an empty key (A14).

`SetTimezone` (#96), which the app sends on every launch, is the first Member write, and a Member write
never gets a minted key: one without the `Idempotency-Key` header is `invalid_argument` before any
service is called, since a key minted per call would make every retry a new write (A14). It creates the
Member through `members.GetOrCreateBySubject`, as `GetHome` does, then hands the key, that Member's id
and the zone to `members.SetTimezone`, and answers the Profile it returns.

`GetOnboardingState` (#98) answers where onboarding resumes: step 1 until the Display Name and Avatar are
saved, step 2 until the Currency is, then, since steps 3 and 4 can be skipped, the highest step saved; 0
once none is left. It carries the Profile so far, with the Currency's decimals `members` adds, and the
`given_name` claim of the caller's own token as `suggested_name`, which fills step 1's First name field: the
email sign-in path never gives the phone an identity token, so the BFF reads the claim. Once the Currency is
saved it also carries the Salary and Saving Target in force from `ledger.GetSettings`, so a resumed step 3 or 4,
or one Home's prompt reopens after setup, is filled in; a failed `ledger` call is `unavailable`. `SaveProfile` checks the answers with `internal/answers`, the
rules `members` applies again, before any service call, then hands them and the step to
`members.SaveProfile` under the phone's key, one write for both (A14). A Currency saved again during
onboarding then goes to `ledger.RelabelCurrency` under the same key, which moves a saved Salary to the new
code without converting it, and the answer carries that Salary (#103); after onboarding a Currency change is
`failed_precondition`, since converting is `ChangeCurrency`'s (PRD §8.9).

`SaveSalary` (#103) is step 3, and later Settings' Salary: `ledger.SaveSalary` under the phone's key, marked
as a setup save while onboarding is unfinished, so the Salary starts in the sign-up month whatever it
changes (spec #92), then, the first time, step 3 recorded in `members` under the same key, after `ledger`
answers. One call and one key for the phone, two writes to two services: a crash between them leaves the
Salary saved and step 3 next, and saving again leaves the newest row in force. `ledger`'s code, such as
`invalid_argument` for an amount with too many decimals or the D56 `failed_precondition`, reaches the phone
as it is.

`SaveTarget` (#104) is step 4, and later Settings' Saving Target, the same way: `ledger.SaveTarget` marked as a
setup save while onboarding is unfinished, then step 4 recorded in `members` under the same key, which ends
onboarding. D44's `failed_precondition`, a Target with no Salary, records no step. After onboarding the first
Salary or Target starts in the current month and later ones follow D56 and D65 (PRD §8.10). Both answer the
whole Home `GetHome` answers, read in the Member's zone with `notify`'s banner, and the Settings view (#170).

`GetSettings` (#170) is Settings' view: the Profile from `members.GetOrCreateBySubject`, then
`ledger.GetSettings`, both required, each with the one-second timeout and its code passed through with no
upstream text. The Profile's Currency code and decimals are replaced by the code `ledger` holds the amounts in,
so no view pairs converted amounts with an older code (spec #169). Notification preferences are not joined yet.

`CancelDeletion` (#105) is "Keep my account" on the restore sheet: `members.CancelDeletion` under the phone's
key, whose `failed_precondition` for a stamp past the 24 hours reaches the phone as it is. The sheet itself is
opened by the stamp in `SetTimezone`'s Profile, the launch call.

Redis holds two of §4.4's limits, through one client. `RequestMagicLink` is counted against its address
before the realm is asked, `ml:<address>:<hour>` with the address lower-cased: five per address per hour
(PRD §11), failing closed with `resource_exhausted`. Every other RPC is counted against the token's subject
once the bearer check has passed, `rl:<subject>:<minute>`: 120 per Member per minute (#97), past which the
request is `resource_exhausted` before any service is called. That one fails open: when Redis does not
answer within a quarter of a second the request goes on uncounted and a log line records it, so a cache
that is down never locks a Member out of their money (§4.4, A6).

Redis also answers a retried Member write (#127), after the per-Member limit has counted it. `SetTimezone`,
`SaveProfile`, `SaveSalary`, `SaveTarget`, `CancelDeletion`, `LogSpend`, `EditSpend` and `DeleteSpend` are
listed in `internal/bff/replay.go`, and a write is covered by being listed there. A successful answer is kept
24 hours under `idem:<subject>:<Idempotency-Key>`, its bytes after the request hash `internal/idempotency`
computes; the same key with the same request gets those bytes back and no service is called, while the same
key with another request goes on to the owning service, which refuses it. A failed answer is never kept,
and the handlers set no response headers, so a replay differs from the first answer in nothing. It fails
open like the per-Member limit: each Redis call has a quarter of a second, and a miss, an error or a timeout
sends the write to its service and logs the degraded call. The owning services' tables stay the guarantee
(A14). `RequestMagicLink` is not covered: it has no Member.

`GetHome` first calls `members.GetOrCreateBySubject` with the token's subject and `email` claim (PRD
§5 step 5, first-sign-in detection), forwarding the caller's token unchanged so `members` verifies
it again, with a one-second timeout; the Profile it answers is the one the HomeView shows. Until the
Member has answered onboarding's steps 1 and 2, a name, an Avatar and a Currency, it stops there with
`failed_precondition`, which the phone reads as "open onboarding" (spec #92): Home greets by name and
prints every amount in that Currency, so neither has a fallback. Then the
§6.1 fan-out runs in parallel with a one-second timeout each: `ledger.GetMonthSummary` (required) and
`notify.GetWarningState` (optional). A failed `members` call reaches the phone with its own Connect
code, a failed `ledger` call as `unavailable`, each with a message naming only the call; the upstream
text, which can hold a host or a database name, goes to the log. `ledger` is the real service since #103,
`notify` since #132, whose kept copy is the banner and whose failure is Home without it. The join
stamps each amount left without a code with the Member's Currency. The month asked for, the month shown and
`first_recap_on` are read in the Profile's zone, the one `ledger` reads "today" in, and `days_remaining` is
`ledger`'s (A2, A3), so a Member in Jakarta just past midnight on the 1st sees the new month. The HomeView also
carries the Currency's decimals from the Profile, so the Log a spend amount field allows exactly that many, and
`ledger`'s earliest Spend date (D72), where the sheet's date picker starts.

`LogSpend` (#129) is ARCHITECTURE §6.2: `ledger.LogSpend` under the phone's key, its Connect code passed
through; for a Spend in the current month, `notify.EvaluateWarning` with ledger's `crossed_line` and copy under
the same key, within 300 ms (#132); then the same join `GetHome` makes, so the answer is the Spend and the
current month's Home even for a Spend dated in a past month. The banner is `EvaluateWarning`'s kept copy when it
answered, and `GetWarningState`'s otherwise, a past-month Spend included, which stamps nothing.

`EditSpend` (#133) runs the same steps with `ledger.EditSpend`, notify told only when the Spend now sits in the
current month. `DeleteSpend` (#133) calls `ledger.DeleteSpend` and never notify, since a delete crosses no line:
its Home's banner is `GetWarningState`'s, so a warning already fired stays after the delete (PRD §8.8).

`ListSpends` (#131) is the Spend list (PRD §6.10, D59) for the month asked, or with none the current month in
the Profile's zone: `ledger.ListSpends`, `ledger.ListMonthsWithData` and `ledger.GetMonthSummary` for the
month's Streak, asked at once with a one-second timeout each. All three are required, and a failed one reaches the
phone with its own code, such as `invalid_argument` for a month that is not `YYYY-MM`. The view carries the
Currency's decimals from the Profile, as the HomeView does.

`GetRecap` (#152) is one month's Recap view (PRD §6.11 to §6.14, §6.17): the Member from `members`, then
`ledger.GetRecap` for the month, each required with a one-second timeout, joined with the Member's current Display
Name and Currency code, so a new name reaches old Recaps too (§7). `ledger`'s codes pass through: `invalid_argument`
for a malformed or later month, `not_found` for a past month with no Spend and no Money In. It is a read: no
`Idempotency-Key` and no replay, and the per-Member limit counts it.

| Path | Holds |
| --- | --- |
| `main.go` | The binary; serves on `BFF_LISTEN` (default `:8080`). No subcommands, and its own start-up rather than `serve.Main` (see below). |
| `../../internal/bff/` | The handlers, the Salary save in `salary.go`, the `Notify` interface, the magic-link path in `magiclink.go`, `LogSpend`, `EditSpend` and `DeleteSpend` beside the other Member writes in `writes.go`, `ListSpends` in `spends.go`, `GetRecap` in `bff.go`, the unit tests with a fake `ledger` served over Connect, and the end-to-end tests (`TestEndToEnd`, `TestMagicLinkEndToEnd`, `TestOnboardingEndToEnd`, `TestLogSpendEndToEnd`, `TestTargetWarningEndToEnd`, `TestGetRecapEndToEnd`, `TestEdge`). |
| `../../internal/auth/` | The JWT interceptor and the token-forwarding client interceptor. |

The BFF keeps its own `main.go` rather than `serve.Main` (`apps/README.md`, "Every service") because
it has none of the three things that body is: no database URL, no pool to open, no schema to apply,
and a `/healthz` with nothing to ping (§6). Putting it
through the same call would mean a nil pool and a nil migration step threaded through every
service's start-up, which is a silent breakage away from a stack that does not come up.

## Environment

| Variable | Default | Meaning |
| --- | --- | --- |
| `BFF_OIDC_ISSUER` | required | The `iss` Keycloak stamps into tokens; Compose derives it from `MMR_HOST` and `KEYCLOAK_HOST_PORT` (`http://localhost:8080/realms/mmr` with the stock `.env`). |
| `BFF_OIDC_JWKS_URL` | required | Where to fetch the signing keys from this process: `http://keycloak:8080/realms/mmr/protocol/openid-connect/certs` inside the Compose network, `http://localhost:8080/...` from the host. The two are separate settings because the phone and a container reach Keycloak by different hostnames. |
| `BFF_OIDC_AUDIENCE` | `bff` | The `aud` a token must carry. |
| `BFF_MEMBERS_URL` | required | The `members` service: `http://members:8080` in Compose. |
| `BFF_LEDGER_URL` | required | The `ledger` service: `http://ledger:8080` in Compose. `GetHome`, `GetOnboardingState`, `SaveProfile`, `SaveSalary`, `SaveTarget`, `LogSpend`, `EditSpend`, `DeleteSpend`, `ListSpends` and `GetRecap` call it, forwarding the caller's token. |
| `BFF_NOTIFY_URL` | required | The `notify` service: `http://notify:8080` in Compose. `RequestMagicLink` calls it, and `GetHome`, `LogSpend`, `EditSpend` and `DeleteSpend` for the Target Warning, forwarding the caller's token. |
| `BFF_REALM_URL` | required | The realm from this process, parent of the token and `login-token` endpoints: `http://keycloak:8080/realms/mmr` in Compose. Separate from `BFF_OIDC_ISSUER` for the same reason `BFF_OIDC_JWKS_URL` is. |
| `BFF_CLIENT_SECRET` | required | The `bff` client's secret, which buys the service-account token the `login-token` endpoint needs. An OpenTofu variable rather than a generated value (§7.5), and the one secret the BFF reads; `BFF_CLIENT_SECRET_FILE` works like every other `_FILE`. |
| `BFF_MAGIC_LINK_BASE_URL` | required | The address the link is built on, the `lt` token its query: `http://<domain>:80/magic` on the machine today and `https://<domain>/magic` once TLS arrives (#87), `http://localhost:8090/magic` on a laptop, where Nginx serves the fallback page on `NGINX_HOST_PORT` (#72). |
| `BFF_REDIS_ADDR` | required | Redis for the two limits and the idempotency fast path: `redis:6379` in Compose, `localhost:REDIS_HOST_PORT` from the host. No default, because a BFF that cannot find Redis refuses every magic-link request (§4.4); the per-Member limit and the fast path let requests through without it. |
| `BFF_LISTEN` | `:8080` | Listen address. |

## Run

The normal way is the whole stack (`infra/compose/README.md`, "Run it"), which starts `bff` once
`members`, `ledger` and Redis are healthy and the realm step has completed, and publishes it on
`BFF_HOST_PORT`.

By hand, against that stack's Keycloak, `members` and `ledger`, on a port the stack does not publish:

```bash
cd apps/backend
BFF_OIDC_ISSUER=http://localhost:8080/realms/mmr BFF_OIDC_JWKS_URL=http://localhost:8080/realms/mmr/protocol/openid-connect/certs BFF_MEMBERS_URL=http://localhost:8081 BFF_LEDGER_URL=http://localhost:8084 BFF_NOTIFY_URL=http://localhost:8083 BFF_REALM_URL=http://localhost:8080/realms/mmr BFF_CLIENT_SECRET="$BFF_CLIENT_SECRET" BFF_MAGIC_LINK_BASE_URL=http://localhost:8090/magic BFF_REDIS_ADDR=localhost:6379 BFF_LISTEN=:8085 go run ./cmd/bff
```

A token for the realm's test Member and the `GetHome` call are in `infra/compose/README.md`, "Try it".

## Tests

`go test ./internal/bff/` runs the unit tests, which need nothing; each test's name and doc comment
say what it proves. The `*EndToEnd` tests and `TestEdge` drive the running stack at
the BFF's HTTP boundary and through the edge, and the fast path's tests use its Redis; their variables are in `infra/compose/README.md`,
"Running the Go tests against this stack".
