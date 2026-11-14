# My Money Recap: Architecture

| | |
|---|---|
| Status | Draft v2, 2026-09-16. Decisions A1–A15 are listed in §11, which is the decision register; its Status column is the only place acceptance is recorded. |
| Scope | The backend behind the iOS app named in `PRD.md` D11: the edge, the API the phone calls, the services, the broker, the cache, the databases and identity. The iOS app's internals live in `IOS.md` (beside this file), whose decisions are numbered `I*`. |
| Vocabulary | `CONTEXT.md` is the glossary. Terms in **bold caps** are defined there. Service names are lower-case code: `members`, `ledger`, `notify`, `currency`, `bff`. |
| Rules | `PRD.md` owns every product rule this document implements. Section references like §8.9 point there. |

## 1. Summary

The iPhone talks to one thing, a Go **BFF** (backend for frontend), through Nginx. The BFF verifies the Keycloak token, applies per-**Member** limits, and joins answers from four Go services: `members` (who the Member is), `ledger` (every money fact and all of PRD §8) and `notify` (every notification in PRD §10), each with its own Postgres database, and `currency` (exchange rates), which has none and keeps the day's rates in Redis. Services never read each other's databases. They talk synchronously over Connect RPC and asynchronously through RabbitMQ using the outbox pattern. Redis holds rate limits, idempotency keys, the daily rate table and the JWKS. Keycloak handles sign-in and owns no app data.

Payment and AI/ML services after phase 1, and web and Android clients later, plug in without changing this shape: a new service is one more box behind the BFF, and a new client is one more generated client of the same contract.

## 2. Decisions this document rests on

| Source | Decision |
|---|---|
| PRD D11, D31, §11 | Native iOS 17+, SwiftUI. Money is printed on the phone by the platform formatter. The phone caches its last snapshot in SwiftData. |
| PRD D6–D8, §9 | Keycloak for identity: Apple, Google and magic-link sign-in. Every service verifies the JWT on every request but the pre-authentication procedures §4.2 point 3 names. |
| PRD §7 | Postgres. Amounts are `NUMERIC` with an ISO 4217 code. Salary and target settings are append-only rows. |
| PRD D61, D63 | Read offline, write online. Offline, the phone shows its last snapshot with a "Last updated" line. |
| A2 (this doc) | The server computes every figure; the phone only stores and displays. |
| A3 (this doc) | One IANA timezone per Member, sent by the app; every "today" in PRD §8 and §10 is evaluated in it. |
| A5–A7 (this doc) | Nginx at the edge, RabbitMQ for events, Redis in phase 1. |
| Product direction | Go services, one container each. Payment and AI/ML services after phase 1. Web and Android clients later. |

## 3. The whole system

```mermaid
flowchart LR
    Phone[iPhone<br/>SwiftUI, SwiftData]
    KC[Keycloak<br/>own db, sends no mail]
    SMTP[SMTP provider]
    NG[Nginx<br/>TLS, per-IP limits,<br/>static AASA + fallback]
    BFF[bff · Go<br/>JWT verify, per-Member limits,<br/>idempotency, screen RPCs, joins]
    R[(Redis)]
    M[members]
    L[ledger]
    N[notify]
    C[currency]
    MQ{{RabbitMQ<br/>exchange mmr.events}}
    PG[(Postgres<br/>four databases)]
    APNS[APNs]
    OER[Open Exchange Rates]

    Phone -. sign-in, PKCE .-> KC
    Phone ==> NG ==> BFF
    BFF -. JWKS .-> KC
    BFF --> R
    BFF -- Connect --> M & L & N & C
    M & L & N --- PG
    M & L & N -- publish via outbox --> MQ
    MQ -- consume --> M & L & N
    C -. daily .-> OER
    C --> R
    N --> APNS
    N --> SMTP
```

Solid arrows are synchronous calls a request waits for. Dotted arrows are background traffic. The bold path is what one tap on the phone travels. Nothing on the right talks to the phone directly.

## 4. Cross-cutting rules

### 4.1 One contract file, every client generated (A1)

Every API here, phone-to-BFF and BFF-to-service, is defined once in `packages/contract/money_recap/v1/*.proto`. Buf generates the Go server stubs and clients, the Swift client for the app, and later the TypeScript and Kotlin clients. `buf breaking` runs in CI against `main` and fails the build on an incompatible change, so a renamed field cannot reach the phone. Event payloads on RabbitMQ are protobuf messages from the same package, so events and RPCs share types. Connect speaks plain HTTP/1.1 and HTTP/2, which is why Nginx proxies it with an ordinary `proxy_pass` and why a web client later needs no gateway change.

```mermaid
flowchart LR
    P[packages/contract/<br/>money_recap/v1/*.proto<br/>RPCs and event messages] -- buf generate --> CI[CI step<br/>buf lint · buf breaking]
    CI --> G[Go: server stubs + clients<br/>bff and every service]
    CI --> S[Swift client<br/>the iOS app]
    CI --> T[TypeScript client<br/>web, later]
    CI --> K[Kotlin client<br/>Android, later]
```

### 4.2 Authentication, at three points

1. **Nginx** does not read the token. Open-source Nginx has no JWT module; that is Nginx Plus. Nginx terminates TLS, limits requests and connections per IP, and hides internal addresses.
2. **The BFF** verifies the Keycloak JWT against the realm's JWKS, cached and refreshed on an unknown key id, on every request but `RequestMagicLink`, its own pre-authentication procedure, exempted by the rule point 3 states, since it is sent before a Member has a token and its handler therefore reads the address from the request rather than from the claims (#69). It rejects expired or wrong-audience tokens before any service is called. A token it cannot check because the JWKS is out of reach is answered `unavailable`, not `unauthenticated`, so the phone keeps its session and its last snapshot instead of refreshing (PRD D61); a token whose key is already cached still verifies, and each service in point 3 answers the same way. It applies the per-Member limit in Redis keyed by the token subject, because thousands of phones behind one carrier address make a per-IP limit useless for fairness.
3. **Each service** verifies the same JWT again, as PRD §9 requires. The BFF forwards the bearer token unchanged. A bug in the BFF therefore cannot expose one Member's rows to another. The reason is that rule's own limit: a pre-authentication procedure is sent before a Member has a token, so it has no Member and no rows to isolate, and it is exempt by procedure name and never by service, leaving every other procedure of the same service behind the check (`internal/auth`). `notify.EmailMagicLink` is one, since the BFF reaches it on a request that carries no token (#68); the BFF's own `RequestMagicLink` is the other (#69).

**A background job has no Member token**, so it sends its own account's token, which a service accepts only for the calls named for that job (#177). Each job is a confidential Keycloak client with a service account and its own secret, so a leaked key opens one door. `notify`'s schedules fetch theirs with the client-credentials grant, cache it until it expires, and send it to `ledger`, which accepts it for `GetSpentToday` and `ListMonthsWithData` alone, without the per-Member owner check: its audience is `ledger`, not `bff`, and every other call, in `ledger` and in every other service, answers it `permission_denied` (`internal/auth`). A Member's own token still reads only its own Member. `members`' deletion job sends its own to Keycloak's admin API alone, holding `manage-users` (§7.1 step 4).

### 4.3 Events: RabbitMQ with the outbox pattern (A4, A5)

In use for the Member's facts: `members` writes `members.created` and `members.updated` to its outbox in the change's own transaction, its relay publishes them, and `ledger` consumes them from `ledger.members` into `member_facts`. `ledger` stages `ledger.spend.logged`, `ledger.spend.edited` and `ledger.spend.deleted` the same way, and `notify` consumes the first two from `notify.ledger` to stamp the Target Warning (§6.2 step 6) and push it once the stamp commits. `notify` consumes `members.created`, `members.updated`, `members.deletion.requested` and `members.deletion.cancelled` from `notify.members` into its own `member_facts`. `internal/events` holds the relay, the consumer and the topology every service shares. The other events and queues below are planned with the services that raise them.

RabbitMQ is a queue: it delivers each message to one consumer in a queue and forgets it after the ack. That is what phase 1 needs. Correctness comes from the pattern around it, which is the same for every broker.

| Element | Setting | Why |
|---|---|---|
| Exchange | `mmr.events`, type topic, durable | One exchange; routing keys name the producer and the fact, e.g. `ledger.spend.logged`. |
| Queues | Quorum queues, one per consumer service and topic family: `ledger.members`, `notify.members`, `notify.ledger`, `members.ledger` (currency converted), `members.purges` (the two `*.member.purged` events, `ledger`'s and `notify`'s; `currency` holds nothing per Member and consumes nothing, #175) | Quorum queues are replicated and are RabbitMQ 4's recommended durable queue. |
| Dead letters | `mmr.dlx` exchange, `<queue>.dead` queues, five tries: `x-delivery-limit` 4, since RabbitMQ counts the failures it forgives; a failure is rejected, never nacked, because since RabbitMQ 4.3 only a reject counts (#101) | A poison message stops after five tries and waits for a human instead of blocking the queue. |
| Publishing | Publisher confirms on; message id = outbox row id; persistent | The relay marks an outbox row sent only after the broker confirms it. |
| Consuming | Manual ack, prefetch 16, `processed_events` table keyed by message id | At-least-once delivery means duplicates happen; the table makes them harmless. |

```mermaid
sequenceDiagram
    participant DB as ledger db
    participant Relay as outbox relay (goroutine in ledger)
    participant MQ as RabbitMQ
    participant N as notify consumer
    Note over DB: one transaction:<br/>INSERT spend (£420, Games, 27 Sep)<br/>INSERT outbox (SpendLogged #8812)
    Relay->>DB: poll every 1 s
    Relay->>MQ: publish #8812 (persistent)
    MQ-->>Relay: confirm → mark outbox row sent
    MQ->>N: deliver #8812
    N->>N: 1. seen #8812 in processed_events? yes → ack, stop
    N->>N: 2. crossed_line in the event and no MonthFlag? write MonthFlag + #8812 in one transaction
    N->>MQ: 3. ack
    N->>N: 4. the stamp was fresh? in the background: INSERT sent_log, then push
    Note over N,MQ: crash before the commit → redeliver → step 2 stamps.<br/>Crash after it → that one push is lost, never sent twice.
```

The Spend and its event are written together, so neither can exist without the other. Delivery is at least once; step 1 makes a second delivery harmless. The consumer always evaluates and stamps for itself; it never depends on a synchronous call having happened first. Debezium is the heavyweight relay that reads the outbox from the write-ahead log; a Go loop polling the outbox table every second is the phase-1 version. The relay selects rows with `FOR UPDATE SKIP LOCKED`, so two replicas of a service polling the same table never contend or double-publish.

### 4.4 Redis: five jobs (A6)

| Job | Owner | Key | TTL | If Redis is down |
|---|---|---|---|---|
| Per-Member rate limit | bff | `rl:{subject}:{minute}`, INCR + EXPIRE | 2 min | Fail open: requests pass, a metric fires (a log line until OpenTelemetry lands, A11). Nginx per-IP limits still stand. With Redis up, the 121st request of a minute is refused with `resource_exhausted` (#97). |
| Magic-link rate limit | bff | `ml:{address}:{hour}`, INCR + EXPIRE | 1 h | **Fail closed**: the request is refused with `resource_exhausted`. PRD §11 makes five per address per hour a security requirement rather than fairness, and no backstop can stand behind it: `limit_req` keys on an IP address and cannot see an email address inside a request body, so §5's limit is not this one. The key is the address, not the subject, because `RequestMagicLink` carries no token. |
| Idempotency of writes, fast path | bff | `idem:{subject}:{Idempotency-Key}` → stored response | 24 h | Fail open. The guarantee is not here: every write RPC carries the key to the owning service, which enforces a unique `(member_id, idempotency_key)` inside the same transaction as the row (A14). A retry after a BFF crash or a Redis restart hits that constraint and gets the original result back, so a retried Save never creates two Spends. Redis only saves the round trip. In use for every Member write: `SetTimezone`, `SaveProfile`, `SaveSalary`, `SaveTarget`, `CancelDeletion`, `LogSpend`, `EditSpend`, `DeleteSpend`, `RequestExport`, `ForgetDevice`, `RequestDeletion`. The stored answer follows the request's hash, so the same key with another request goes on to the owning service, which refuses it; a failed answer is never stored. `RequestMagicLink` is not covered. |
| Exchange-rate cache | currency | `fx:{yyyy-mm-dd}:USD` → the day's rate table | 36 h | Fetch from OER again, one of the free plan's 1,000 a month; there is no snapshot table. |
| JWKS cache | bff, services | `jwks:{realm}` | 1 h | Each process also keeps the last JWKS in memory, so token checks continue. |

Redis 8 is licensed AGPLv3 (alongside RSALv2 and SSPLv1). AGPL binds Redis itself, not programs that talk to it over a socket; nothing here modifies Redis. Valkey (BSD-3) is a drop-in swap if a licence policy ever requires it.

### 4.5 Time: one zone per Member, no zone list in code (A3)

The app sends `TimeZone.current.identifier` on every launch. `members` validates it with Go's embedded IANA database (`time/tzdata`) and stores it. Every "today" in PRD §8 and §10 is evaluated in that zone by the service doing the arithmetic: Money In on or after Payday (§8.2), the streak window (§8.5), days remaining (§8.7), "never in the future" on a Spend date (§8.10), and the Recap-ready and daily-reminder triggers (§10). Spend dates stay plain calendar dates, as D34 says. So is the sign-up date: `members` stores `signed_up_on`, the local date at the moment of sign-up, next to `created_at`, and the streak window (§8.5) and the partial-month rule (§8.11) read that date. A Member who later moves to another zone changes only what "today" means from now on; no past month is recomputed, because nothing in a past month reads the zone.

The `notify` scheduler does not compute local time per row on every tick. Each `member_facts` row carries `next_recap_at` and `next_reminder_at` as UTC timestamps, computed with `AT TIME ZONE` by two SQL functions, `next_recap` and `next_reminder`, when the zone or the preferences change and again after each fire, and NULL while that push's switch is off; both columns are indexed. Each schedule's tick is then one indexed range query:

```sql
SELECT member_id FROM member_facts
WHERE next_reminder_at <= now()
ORDER BY next_reminder_at
LIMIT 500;
```

No cron per zone and no zone table. Postgres and Go both carry the IANA data, so a daylight-saving change is a package update, and the next-fire columns are recomputed from it.

### 4.6 Observability (A11)

Once one tap on the phone crosses Nginx, the BFF and up to three services, a slow or failed request cannot be understood from one log file. Every Go process, the BFF included, uses the OpenTelemetry SDK for traces, metrics and logs and exports them over OTLP to one container, `grafana/otel-lgtm`, which bundles Tempo (traces), Loki (logs), Prometheus (metrics) and Grafana in a single image for phase 1. The BFF starts a trace per request and propagates the context in Connect headers and in RabbitMQ message headers, so a `SpendLogged` event and the push it caused appear under the Save that produced them. Every log line carries the trace id and the Member id, never the token. The three alerts worth having on day one: any service health check failing, any queue's dead-letter count above zero, and BFF p95 latency above one second (PRD §11).

## 5. The edge: Nginx (A7)

| | |
|---|---|
| Responsibility | TLS termination, per-IP and per-connection limits, routing, static files, hiding every internal address. No business logic, no token parsing. |
| Routes | `/api/` → bff. `/auth/` → Keycloak. `/.well-known/apple-app-site-association` → static, `application/json`, no redirect. `/magic` → the static fallback page. |
| Limits | `limit_req_zone $binary_remote_addr rate=60r/s burst=120`, deliberately generous because a mobile carrier can put thousands of phones behind one address. `limit_conn` 100 per address. A stricter 5 r/min on the magic-link path only, which blunts a flood from one IP address and is **not** PRD §11's five per email address per hour: `limit_req` keys on an IP address and cannot see an address inside a request body, so the BFF enforces that limit itself, fail closed (§4.4). |
| Certificates | Let's Encrypt via an `acme.sh` sidecar whose deploy hook runs `nginx -s reload` (`certbot/certbot` was rejected: its docs say it cannot install a certificate or reload a web server). |
| Why Nginx | The most widely deployed reverse proxy, documented everywhere, and it keeps TLS and abuse control out of Go. |
| Known limit | JWT validation needs Nginx Plus or the njs module. The design does not depend on it; the BFF verifies tokens. |

## 6. The API: `bff`

| | |
|---|---|
| Responsibility | The only thing the phone talks to. Screen-shaped RPCs, JWT verification, per-Member limits, idempotency, fan-out to services and the join. Owns no database. |
| RPCs | `RequestMagicLink`, `GetHome`, `GetRecap(month)`, `ListSpends(month)`, `LogSpend`, `EditSpend`, `DeleteSpend`, `GetOnboardingState`, `SaveProfile`, `SetTimezone` (called on every launch), `GetSettings`, `SaveSalary`, `SaveTarget`, `PreviewCurrencyChange`, `ChangeCurrency`, `SaveNotificationPrefs`, `RegisterDevice`, `ForgetDevice`, `RequestExport`, `RequestDeletion`, `CancelDeletion`. Each is a thin join or pass-through to the service RPCs in §7. |
| Depends on | `members`, `ledger`, `notify`, `currency` over Connect; Redis; Keycloak JWKS. |
| Publishes | Nothing. The BFF never writes state, so it needs no outbox. |
| Why a BFF | Home needs the profile (`members`), the numbers (`ledger`) and the warning state (`notify`). Someone must join them. Doing it in Go keeps the phone a pure display and keeps the join out of Nginx, which cannot do it. |

### 6.1 Flow: GetHome

```mermaid
sequenceDiagram
    participant P as iPhone
    participant B as bff
    participant M as members
    participant L as ledger
    participant N as notify
    P->>B: GetHome (bearer token)
    B->>B: verify JWT, per-Member limit
    B->>M: GetOrCreateBySubject → Profile, 1 s timeout
    par in parallel, 1 s timeout each
        B->>L: GetMonthSummary(current)
        B->>N: GetWarningState(current)
    end
    B->>B: join → HomeView
    B-->>P: HomeView; phone stores it in SwiftData
```

Before the fan-out, a Member who has not answered onboarding's steps 1 and 2 gets `failed_precondition`, which the phone reads as "open onboarding" (spec #92, #100): Home greets by name and prints every amount in the Member's Currency, so it has no fallback for either. `members` and `ledger` are required: if either times out, the BFF returns an error and the phone keeps drawing its last snapshot. `notify` is optional: if it times out, the HomeView is returned without a warning banner and a metric fires. A notification service can never blank the Home screen.

### 6.2 Flow: LogSpend

1. The phone sends `LogSpend` with an `Idempotency-Key`, one per save attempt: a retry of the same values sends the same key, and changed values a new one (IOS I3). Every write RPC in this document carries one; `LogSpend` is the worked example.
2. The BFF verifies the JWT, checks the per-Member limit, and looks up the key in Redis. A hit returns the stored response without touching `ledger`.
3. The BFF calls `ledger.LogSpend(key, …)`. Ledger looks up `member_facts`; on a miss (the Spend arrived before `members.created` was consumed) it calls `members.GetProfile` once and upserts, and if that fails it returns a retryable error. It validates D57, D72 and §8.10 in the Member's zone, then in one transaction writes the Spend with the key under the unique `(member_id, idempotency_key)`, the outbox row, and recomputes the month. A duplicate key returns the original Spend's `MonthSummary`, not an error. The response includes `crossed_line`, computed here against the §8.8 line, so no other service does that arithmetic.
4. The BFF calls `notify.EvaluateWarning(month, crossed_line, warning)` with a 300 ms timeout, `warning` being ledger's §8.8 copy. Notify stamps `month_flag` with that copy if `crossed_line` and no stamp exists, and returns the warning state. On a timeout or a failure the BFF continues and reads the banner from `GetWarningState(current)` as in §6.1; the event path below will stamp the month a second later. `notify` can be down and a Member can still record a coffee.
5. The BFF stores the response under the idempotency key for 24 hours and returns a fresh `HomeView`. The phone animates the numbers (§6.9) and replaces its snapshot.
6. Separately, the relay publishes `ledger.spend.logged` carrying the same figures and `crossed_line`. Notify consumes it and stamps `month_flag` if not yet stamped. Whichever of this step and step 4 stamps the month first pushes it, in the background after the stamp commits (§7.3). The stamp is idempotent on `(member_id, month)` and the push on `sent_log`, so it does not matter whether step 4 ran, ran twice, or never ran. The event path is the guarantee; the synchronous call is only the user experience.

`EditSpend` runs the same six steps with `EditSpendResponse` and `ledger.spend.edited`. Step 4 names the month the Spend now sits in (`EditSpendResponse.month`), the only month an edit can push over the line; the month it left only loses Spent, which never un-fires a warning (PRD §8.8). Ledger answers `crossed_line` true only when the write's month is the current one (§8.8), and the BFF skips step 4 for a past month, so it stamps nothing. Home is always the current month, so when a write's month is a past one (a Spend logged, edited or moved there), the fresh `HomeView`'s banner comes from `GetWarningState(current)` as in §6.1, never from that month's kept copy. `DeleteSpend` never crosses the line (`DeleteSpendResponse.crossed_line` is always false), so it skips step 4 and its banner comes from `GetWarningState(current)` too.

## 7. The services

### 7.1 `members`

| | |
|---|---|
| Responsibility | Who the Member is: Keycloak subject, email as the token's `email` claim reports it, Display Name, Avatar, Currency code, timezone, sign-up date, deletion request. First-sign-in detection (PRD §5 step 5). Onboarding progress (§6.4). The 24-hour deletion job (§9). |
| Owns | `member` (id, keycloak_subject, email, display_name, avatar_key, currency_code, currency_at (the newest conversion applied), timezone, created_at, signed_up_on, deletion_requested_at, deleted_at, onboarding_step), `deletion_progress` (member_id, keycloak_deleted_at, ledger_purged_at, notify_purged_at, currency_purged_at), `outbox`, `processed_events`. |
| RPCs | `GetOrCreateBySubject` (saves the token's `email` claim on create and refreshes it on every call whose token carries one, so a change made in Keycloak reaches us on the next sign-in, and publishing `members.updated` when it changed (§4.3); a refreshed token lacks the claim after an admin clears the address, and the stored one stays), `GetProfile`, `SaveProfile` (Display Name, Avatar, Currency and the onboarding step just completed or skipped, one write per step), `SetTimezone`, `RequestDeletion`, `CancelDeletion`. Every write takes an idempotency key (A14). |
| Publishes | `members.created`, `members.updated` (its `MemberFacts`: name, timezone, sign-up date, currency, email; sent only when one of them changed, so an Avatar-only save sends none), `members.deletion.requested`, `members.deletion.cancelled`, `members.deleted`. |
| Consumes | `ledger.currency.converted` → set `currency_code`, then publish `members.updated`. `ledger.member.purged` and `notify.member.purged` → stamp `deletion_progress`. |
| Why separate | Identity and profile change rarely and are needed by every other service. Publishing them as events lets `ledger` and `notify` keep the few facts they need, timezone, `signed_up_on` and email, without calling `members` on every request. The email is an identifier, not a credential: it is the standard OIDC claim, read from a signed token over TLS, and storing it carries the same risk as storing the Display Name. |

Flow: deletion and restore (D39, D60)

1. `RequestDeletion` stamps `deletion_requested_at` and stages `members.deletion.requested` in one transaction under the Member's key; a repeat keeps the first stamp. The BFF answers the stamp, and the phone signs out as Log out does: it forgets this phone in `notify`, ends the realm session and erases its cache. `notify`'s deletion-pending flag, set by the event, holds back every push to the Member's other phones.
2. A sign-in within 24 hours: the launch call (`SetTimezone`) answers the Profile with the stamp, and when it is under 24 hours old the phone shows the restore sheet (PRD §6.18) before the setup check and before Home, even for a Member the stored Home greets. "Keep my account" calls `CancelDeletion`, which clears the stamp and publishes `members.deletion.cancelled` in one transaction; a stamp 24 hours old or more is `failed_precondition`, and with no stamp nothing changes. "Continue with deletion" is the phone's own sign-out. The job (steps 3 to 6) is planned (#184).
3. A worker in `members` runs every 10 minutes. For every stamp older than 24 hours it sets `deleted_at` on the member row, inserts a `deletion_progress` row, and writes the `members.deleted` outbox row, all in one transaction. The row is a tombstone: `GetOrCreateBySubject` treats it as absent. No remote call is inside this transaction.
4. A second step, retried until it succeeds, deletes the Keycloak user through the admin API with `members`' own service account (§4.2, built in #177 with no job calling it yet) and stamps `keycloak_deleted_at`. A 404 from Keycloak counts as success, so a retry after a half-finished run is harmless.
5. `ledger` and `notify` each consume `members.deleted`, purge that Member's rows idempotently, and publish `<service>.member.purged`. `members` consumes those two and stamps `deletion_progress`. `currency` stores nothing about any Member and takes no part (#175).
6. When all three stamps are present the member row is hard-deleted. An alert fires for any `deletion_progress` row incomplete one hour after `deleted_at`, which is how a broken consumer is noticed instead of leaving Spend rows behind indefinitely.

### 7.2 `ledger`

Built in part: the Salary and Saving Target rows, `GetSettings`, `GetMonthSummary` and the replica, kept by `members.created` and `members.updated` and filled on a miss. The BFF calls it for Home's month and onboarding's steps 3 and 4, which save with `setup` while onboarding is unfinished and relabel on a Currency change (`RelabelCurrency`). A setup save starts in the sign-up month, a first save after onboarding in the current month (PRD §8.10). It stores Spends (`LogSpend`, `EditSpend`, `DeleteSpend`, `ListSpends`), counts them in every Home figure, `crossed_line` included, and stages each Spend write's event in the outbox. It answers `GetRecap` for every Recap screen and Home's past-Recap list: a Recap names its Streak run's first and last day and its Archetype's title, and is `not_found` for a past month Home does not list. It answers `GetSpentToday`, and it and `ListMonthsWithData` accept `notify`'s own token (§4.2). It previews and applies the currency change (`PreviewConversion`, `ApplyCurrencyConversion`) and stages `ledger.currency.converted`; the BFF's Home, Recap, Spend list and Settings take their Currency from its answers. It answers `ExportCSV`. Planned: the other events it publishes and the `members.deleted` purge.

| | |
|---|---|
| Responsibility | All money facts and **all** of PRD §8, nothing of it anywhere else: salary and target settings, Spends, currency rewrites and their rounding (§8.4, §8.9), and every computed figure: Spent, Money In, Saved, target amount, kept%, Streak, Splurge, Archetype, partial-month status, the target sentence, the warning line and whether it was crossed (§8.8). `notify` receives a boolean; `currency` supplies a rate. |
| Owns | `salary_setting`, `saving_target_setting`, `spend`, `currency_conversion`, `member_facts` (replica: member_id, keycloak_subject, timezone, signed_up_on, currency_code, facts_at, currency_at; events own the facts, `facts_at` being the newest applied event's time, so an older one arriving late is skipped, and `currency_at` the last `RelabelCurrency`'s or conversion's, so an event older than it cannot take its Currency back, and the profile read on a miss fills only facts the row lacks; the subject is the one whose token the profile read succeeded with, and a call under another token is a miss, so `members` answers it, §4.2 point 3), `idempotency`, `outbox`, `processed_events`. `spend` and `salary_setting` carry `original_amount` and `original_currency`, the values as typed, untouched by a later conversion; every write claims its key in `idempotency` under a primary key on `(member_id, idempotency_key)` inside the write's transaction, as `members` does (A14; one table rather than a key column per write table). |
| RPCs | `GetMonthSummary(month)`, `GetRecap(month)`, `ListMonthsWithData`, `ListSpends(month)`, `LogSpend`, `EditSpend`, `DeleteSpend`, `GetSettings`, `SaveSalary(applies_from, setup)`, `RelabelCurrency(to)`, `SaveTarget(setup)`, `PreviewConversion(rate, to)`, `ApplyCurrencyConversion(rate, source, to)`, `GetSpentToday(member_ids[])`, `ExportCSV`. Every write takes an idempotency key (A14). |
| Rounding | Decimals per currency come from Unicode CLDR's `currencyData.json`, release 48.2.0, kept unedited with its licence in `internal/cldr`, the way Go's `time/tzdata` carries IANA's zones, so no table of currencies is written by hand. A code is a Currency when some territory takes it, or once took it, as tender, and its decimals are the file's: GBP 2, IDR 0, BHD 3. Step 2's check (`internal/answers`) and `ledger`'s rounding read that one lookup. The phone prints an amount with exactly the decimals it carries, never its own count for the currency, which differs between iOS versions (the forint has 2 on iOS 26.0 and 0 on 27.0); a test in the app checks that every currency it lists is one the file accepts. A new CLDR release is a deliberate bump: replace the file, change the checksum its test holds, and check whether any Member's Currency changed decimals. The stored value is the truth; the phone only prints it. `golang.org/x/text/currency` before it was CLDR 32 from 2017, which refused five current currencies and rounded seven unlike the phone. |
| Publishes | `ledger.spend.logged`, `ledger.spend.edited`, `ledger.spend.deleted`, `ledger.salary.changed`, `ledger.target.changed`, `ledger.currency.converted`, `ledger.member.purged`. Every spend event carries the recomputed month figures and `crossed_line`, so `notify` never recalculates. |
| Consumes | `members.created`, `members.updated` → upsert `member_facts` (queue `ledger.members`). `members.deleted` → purge, then publish `ledger.member.purged`. |
| Tests | The PRD §8 worked examples and the §12 acceptance criteria live as JSON fixtures in `packages/contract/fixtures/ledger/`, typed by the contract's `LedgerFixture`: decoded against the schema by `apps/backend/internal/ledger/fixtures_test.go`; every one runs through `ledger` (`TestTheAnswerKey`), none left out by name, the currency change (`prd-8.9-maya-gbp-to-idr`, PRD §12's BHD to IDR, and a second conversion from converted amounts) included. Each fixture is one ledger and one "as of" date with the expected numbers, including `crossed_line` and the rounded conversion results. The ledger's test suite runs every fixture. Because no other service implements any of §8, this is the single answer key for the arithmetic. |
| Why one service for all of it | §8.9 rewrites salary rows and Spend rows "in one transaction". Keeping both tables in one database keeps that a real transaction instead of a saga. The §8 rules read every table here; splitting them would turn every computation into three network calls. |

Flow: currency change (§8.9)

```mermaid
sequenceDiagram
    participant P as iPhone
    participant B as bff
    participant C as currency
    participant L as ledger
    participant M as members
    P->>B: PreviewCurrencyChange(IDR)
    B->>C: Quote(GBP→IDR)
    C-->>B: rate 20,150, source PROVIDER
    B->>L: PreviewConversion(20,150, IDR)
    L-->>B: before/after values, rounded by ledger
    B-->>P: values for the confirm sheet
    P->>B: ChangeCurrency(IDR, 20,150, key)
    B->>L: ApplyCurrencyConversion(20,150, PROVIDER, IDR, key)
    Note over L: one transaction: rewrite salary + spends<br/>(originals kept), currency_conversion row, outbox row
    L--)M: ledger.currency.converted (via RabbitMQ)
    Note over M: set currency_code, publish members.updated
```

The rewrite of every amount is one Postgres transaction inside `ledger`; each rewritten row keeps `original_amount` and `original_currency`, so what the Member typed is never lost (PRD §8.9). The same transaction stamps the replica's Currency as `RelabelCurrency` does. `members` learns the new code from the event, not from a call, on its queue `members.ledger` once per message id, and skips a conversion older than the one it holds (`member.currency_at`), so a failure there is retried by the broker and the worst case is a code that lags the amounts by seconds, never amounts that are half converted. No screen shows that lag: every view the BFF answers takes its Currency and decimals from `ledger`'s answer, so the conversion's own answer is in the new Currency. A second tap with the same key returns the first result. A rate is above zero with at most 11 digits before the separator and 12 after, and a rewrite that would pass D57's 11 digits is `out_of_range` with nothing written.

With nothing to convert (§8.9 step 2), `ChangeCurrency` carries no rate and makes onboarding's two writes under one key, `members.SaveProfile` and then `ledger.RelabelCurrency`, so both hold the new code before the answer. Without a typed rate the preview asks `currency.Quote` and hands its rate to `ledger.PreviewConversion` as `PROVIDER` with its date; an unsupported pair answers `unsupported_pair`, and a failed fetch, or no answer, `unavailable` (#175). The BFF does not wait for `currency` to start.

Flow: salary edit after Payday (D56)

1. The phone calls `GetSettings`; the response includes `payday_passed_this_month`, computed by `ledger` in the Member's zone.
2. If true, the phone shows the D56 sheet and sends `SaveSalary(applies_from = this_month | next_month)`. If false, it sends `unspecified` with no sheet; `ledger` reads that as `this_month` before its own Payday clock and answers `failed_precondition` on or after it, so a flag that went stale while Settings was open makes the phone show the sheet instead of rewriting a paid month.
3. `ledger` appends a `salary_setting` row with that `effective_from`. Nothing is updated in place, so September's Recap is reproducible from the rows. The exceptions are the currency change's: a conversion rewrites every amount (§8.9), and `RelabelCurrency` rewrites the code and the rounding of the Salary rows in place while onboarding is unfinished (spec #92), and later only with no Salary and no Spends to move.

### 7.3 `notify`

Built in part: `EmailMagicLink`; the month flag with `GetWarningState` and `EvaluateWarning`, stamped from `ledger.spend.logged` and `ledger.spend.edited` on `notify.ledger`; the replica row (the owner's subject, filled on a miss, and the zone, email and deletion-pending flag from `notify.members`); its own service account and a `ledger` client that sends that account's token (§4.2), which the schedules call; `RegisterDevice`, upserting the `device` row on its token, and `ForgetDevice`, deleting this Member's row for a token; `GetPrefs` (the Member's `notification_preference` row or PRD §10's defaults) and `SavePrefs`, replacing it whole. The first stamp of a month, by either path, pushes the Target Warning through the Apple sender once it commits, claiming `sent_log` first. `EmailExport` mails the CSV. Recap ready and the daily reminder run on the two schedules below. The `members.deleted` purge is planned.

| | |
|---|---|
| Responsibility | Deciding and sending every notification in PRD §10, and remembering what was sent. Device tokens. Notification preferences. The Target Warning state the Home banner reads. All outbound email: the magic-link email (§7.5) and the CSV export (PRD §6.16). Keycloak sends no mail. |
| Owns | `notification_preference` (member_id, recap_ready, daily_reminder, daily_reminder_time, target_warning: one row per Member), `device` (apns_token, member_id, registered_at: one row per phone, upserted on the token so a phone another Member signs in on moves to them), `month_flag` (member_id, year_month, target_warning_sent_at, target_warning_message: ledger's §8.8 copy as it read when the month was stamped, so the Home banner stays until the 1st however Spent moves afterwards), `sent_log` (member_id, kind, period, sent_at: primary key on the first three, claimed before a send so a crash or a second replica never sends twice; kind `target_warning` with the month as its period), `email_status` (member_id, address, last_bounce_at; planned with the provider's webhook, #87), `magic_link_send` (idempotency_key, email, sent_at: one row per magic-link send, unique on `(email, idempotency_key)`: A14's `(member_id, idempotency_key)` has no Member on this path, and the key alone let one address lock every other one out of its own link; every send sweeps rows past the retention window in the same transaction, because a row holding a raw address and no member_id is one the purge below cannot reach), `member_facts` (replica: keycloak_subject, the subject whose token `members.GetProfile` last answered for the Member, so a call under another token is a miss `members` answers, §4.2 point 3; timezone and email with their stamp facts_at, and deletion_pending with its stamp deletion_at, so an older event never overrides a newer one of its kind; a row lacking the zone or the email is a miss too, and the profile read fills both unless a newer event has; no display_name, since no push prints one; and next_recap_at and next_reminder_at, the schedules' fire times), `idempotency` (member_id, idempotency_key, request_hash, response, created_at: A14's table, as in `members` and `ledger`), `processed_events`. |
| RPCs | `GetWarningState(month)` answering `fired` and the kept copy, `EvaluateWarning(month, crossed_line, warning)` answering `fired`, `fired_now` and the kept copy, `GetPrefs`, `SavePrefs`, `RegisterDevice`, `ForgetDevice`, `EmailMagicLink(email, link)`, `EmailExport(csv)`. Every write takes an idempotency key (A14). `notify` holds no rule from PRD §8; `crossed_line` is computed by `ledger`. |
| Email | One SMTP provider, credentials mounted as a secret file (A13), is planned (#87 item 8): today `notify` sends to Mailpit on a laptop and in CI, and the machine has no mail host. `EmailMagicLink` takes the address the BFF parsed, because no member row may exist yet; `EmailExport` reads the address from `member_facts`, and a permanent 5xx refusal of it answers `failed_precondition` with `EmailBounced`; Profile cannot change the address, so the phone's line has no link. Planned with the provider (#87): a later bounce reported by its webhook stamps `email_status.last_bounce_at`. |
| Consumes | `ledger.spend.logged`, `ledger.spend.edited` → stamp `month_flag` if `crossed_line` and unstamped, then, once that commits, push if the stamp is fresh. `members.created`, `members.updated` → replica's zone and email, and a new zone recomputes `next_*_at`. `members.deletion.requested`, `members.deletion.cancelled` → set or clear the replica's deletion-pending flag. `members.deleted` → purge, then publish `notify.member.purged`. |
| Schedulers | Every minute, one indexed query on `next_recap_at`, one on `next_reminder_at` (§4.5), in batches of 500. Recap ready asks `ledger.ListMonthsWithData` for each Member of the batch, the reminder one `ledger.GetSpentToday(member_ids[])` for the whole batch, both with `notify`'s own token; then the pushes, eight Members at a time. Each push inserts into `sent_log` first and sends only if the insert won, so a crash between the two never double-sends and two `notify` replicas never both send; the Member's time then advances. A failed `ledger` call leaves the time as it is, so the next tick tries again; a Member `ledger` does not know moves on unsent. |
| APNs | The Apple sender: token-based auth, an ES256 token signed with the .p8 key, its Key ID as `kid` and the Team ID as issuer, reused for 40 minutes; one HTTP/2 request per device carrying the alert body, to the host and topic from configuration. A 410 response deletes the device row. With no key, as on every laptop, a push is recorded in `sent_log` and skipped with one log line. No push goes out while the Member's switch for it is off or a deletion is pending. The Target Warning's copy is "You've passed your saving line for {Month}.", Recap ready's "Your {Month} recap is ready. See how you did.", each month in English, and the reminder's "Log today's spends before you forget." |
| Why separate | It is the only service with schedulers and an outbound push connection. Keeping it out of `ledger` keeps the money service free of clocks and Apple. |

Flow: Recap ready at 09:00 local

```mermaid
flowchart LR
    T[scheduler tick<br/>every 60 s] --> Q[indexed query<br/>next_recap_at <= now, batch 500<br/>Jakarta rows fire at 02:00 UTC, London at 08:00 UTC]
    Q --> E[ledger.ListMonthsWithData, notify's key<br/>previous month present?]
    E --> I[INSERT sent_log member+month<br/>conflict? skip: already sent]
    I --> A[APNs, eight Members at a time<br/>then advance next_recap_at]
```

A partial month still sends (PRD §10). The daily reminder is the same loop on `next_reminder_at` with one batched `ledger.GetSpentToday(member_ids[])` per 500 Members. At phase-1 scale this is a handful of rows a minute; the batch, the index and the worker pool are what keep a popular reminder time from becoming a burst of one RPC and one push per Member from a single loop.

Flow: CSV export (PRD §6.16, D39)

1. The phone calls `RequestExport` under its Idempotency-Key. The BFF calls `ledger.ExportCSV`, which returns every non-deleted Spend for that Member, oldest first, as one CSV document (RFC 4180, a leading apostrophe on any cell starting with `=`, `+`, `-` or `@`, only the header with no Spends).
2. The BFF calls `notify.EmailExport(csv)` under the same key; both calls are required. Notify reads the address from `member_facts`, claims the key in `idempotency` and sends through SMTP in that transaction, bounded as the magic link's send is, so a failed send frees the key and a retried key sends once.
3. The BFF returns success and the phone's button becomes `Sent ✓` for 60 seconds (D67). A permanent refusal of the address returns `failed_precondition` with `EmailBounced`, which the phone shows as "We can't reach {email}." with no link; any other failure is `unavailable`, "We couldn't send it. Try again." The file is never stored server-side; a second tap regenerates it. Bounces the mail host reports later, and the `email_status` table that keeps them, wait for the provider's webhook (#87).

### 7.4 `currency`

Built: `cmd/currency` and `internal/currency`.

| | |
|---|---|
| Responsibility | Exchange rates: fetching Open Exchange Rates' day's table (D35), caching it, and naming an unsupported pair. A manual rate is typed on the phone and never passes through here (PRD §8.9 step 1); the confirm sheet's before/after figures are `ledger.PreviewConversion`'s. |
| Owns | Nothing in Postgres: no database, no broker, no outbox and no consumer (owner, 2026-09-29). The day's table lives in Redis only. |
| RPCs | `Quote(from, to)` → the cross rate from the USD table, rounded to 8 significant digits, source `PROVIDER` and the table's date; `failed_precondition` with `UnsupportedPair` for a code the table lacks; `unavailable` when the fetch fails or no App ID is configured, as on every laptop stack. No arithmetic on amounts: rounding and rewriting are `ledger`'s (§7.2), so §8 has one implementation. |
| Consumes | Nothing. It holds nothing about any Member, so it takes no part in deletion (§7.1). |
| Redis | `fx:{date}:USD`, 36-hour TTL. A miss, or Redis down, fetches OER again. |
| Why separate | It is the only service that calls a third-party API with a key and a quota. A quota failure or an OER outage returns an `InlineError` with Retry (§8.9 step 1) and nothing else notices. Idle it measured 9.2 MiB on a laptop stack (2026-09-30), under its siblings' 13 to 17. |

### 7.5 Keycloak

| | |
|---|---|
| Responsibility | Sign-in. Brokers Apple and Google, issues login tokens through the Phase Two extension, issues and refreshes session tokens, ends sessions. Holds no app data and sends no email. |
| Owns | Its own database, `keycloak`, in the same Postgres container. |
| Image | `quay.io/phasetwo/phasetwo-keycloak`, pinned to a tag whose bundled `keycloak-magic-link` and Apple identity-provider extensions match its Keycloak version. Never `latest`: the magic-link extension is compiled against one Keycloak minor and has broken on the next. |
| Configuration as code | The official `keycloak/keycloak` OpenTofu provider under `infra/tofu/realm/` (A13): the realm, the `ios` public client, the BFF's service-account client, the `notify` and `members` service-account clients of the two background jobs (§4.2), the `login-token` scope, the custom direct-grant flow, the test Member and the Google and Apple identity providers (#71), whose client ids, Team ID and Key ID are variables and whose secrets are vault references. A one-shot container applies it on every start, so a realm edit takes effect with `docker compose up`. No SMTP block: Keycloak sends nothing. The JSON import it replaced was skipped once the realm existed, so it could not carry later changes; kcadm is imperative. |
| Apple sign-in | The bundled Apple identity-provider extension takes the Team ID, Key ID and `.p8` key and signs the client-secret JWT itself on every code exchange (§9). |
| Session lifetime | Every sign-in asks for `offline_access`, so the app holds an offline refresh token, which the realm keeps for 30 days from its last renewal, the last time the app was used online, with no overall cap (`offline_session_idle_timeout`, PRD D62). The SSO session's defaults, 30 idle minutes and 10 hours, would end it the same day. Each refresh token works once (`revoke_refresh_token`, no reuse): a reused one is refused and ends that sign-in, so a copied token dies at the app's next renewal. Log out sends the current refresh token to the logout endpoint, which ends the offline session as well. |
| Secrets | Keycloak's file vault: `--vault=file --vault-dir=/run/secrets` with the `KEY_ONLY` key resolver, and the two identity-provider secrets referenced as `${vault.google_client_secret}` and `${vault.apple_sign_in_key}`. The files are mounted from the SOPS-decrypted set (§10, A13), under the names Keycloak 26.6 looks for, with every underscore doubled (`infra/secrets/README.md`). Google's resolves. The Apple extension pinned in the image (1.17.0) never asks the vault: it base64-decodes the literal reference as its `.p8`, so Sign in with Apple cannot finish a code exchange with real values until that changes (#71). The exceptions are the three client secrets, `bff`'s, `notify`'s and `members`', each an OpenTofu variable and so in OpenTofu state and in the environment: the file vault resolves its placeholder only for the SMTP password, identity-provider secrets and LDAP, and the BFF (`BFF_CLIENT_SECRET`, for the Login Token endpoint), `notify` and `members` (`NOTIFY_CLIENT_SECRET`, `MEMBERS_CLIENT_SECRET`, §4.2), the A10 smoke script and the tests must know the values in advance. |

Flow: email sign-in (D7, D32, acceptance 2)

Keycloak's magic-link action token cannot be placed on the app's domain and is consumed by a plain GET, which mail scanners issue. The extension's newer **Login Token** path avoids both, so that is A10.

1. The phone sends `RequestMagicLink(email)`. The BFF enforces 5 requests per email per hour in Redis (PRD §11), then calls the extension's `POST /realms/mmr/login-token` with `expiration_seconds: 900`, `reusable: false`, `force_create: true` and `set_email_verified: true`, using its service account. The response is an opaque `lt:<uuid>`.
2. The BFF calls `notify.EmailMagicLink(email, link)` and `notify` emails `https://app.example.com/magic?lt=…`. Nginx serves a static page for `/magic` and the apple-app-site-association file from `/.well-known/`, so on an iPhone with the app installed iOS opens the app instead of the page.
3. The app POSTs `grant_type=password&scope=login-token offline_access&login_token=…` to Keycloak's token endpoint and receives its tokens, the refresh token an offline one ("Session lifetime" above). The `login-token` scope routes the `ios` client's direct grant to the Login Token branch; without it the same flow runs the username-password validators (the test seam), because the Login Token authenticator fails the flow outright when the field is absent, so the two cannot be plain alternatives. No browser is involved on the email path; Apple and Google still use `ASWebAuthenticationSession` with PKCE. An expired or used token returns `invalid_grant`, which the app shows as the §6.3 `InlineError`.
4. On a laptop, or when the Member has set links to open in the browser (Apple TN3155), the static page reads "Open My Money Recap on your phone to finish signing in." and offers no control, as `SCREEN-STATES.md` A3 specifies (`infra/nginx/html/magic.html`, #72). The page performs no request with `lt`, and a scanner GET on the URL reaches only that static page, so the token cannot be consumed by anything but the app.

**Open item, resolved 2026-09-16.** `infra/compose/a10-smoke.sh` proves on `quay.io/phasetwo/phasetwo-keycloak:26.6.6.1789466980` (Keycloak 26.6.6, magic-link 0.75) that the Login Token endpoint issues a token for the test user and the `ios` client redeems it through the direct grant, that the second redemption is refused with `invalid_grant`, and that the password grant still works. CI's `backend` job runs it on every pull request that touches the backend, `infra/` or the contract (#14).

## 8. Data ownership

| Table | Database | Writer | Readers |
|---|---|---|---|
| `member`, `deletion_progress` | members | members | members only; others hold `member_facts` replicas fed by events |
| `salary_setting`, `saving_target_setting`, `spend`, `currency_conversion` | ledger | ledger | ledger only |
| `member_facts` (replica) | ledger, notify | that service's consumer | that service |
| `notification_preference`, `device`, `month_flag`, `sent_log`, `email_status`, `magic_link_send` | notify | notify | notify only |
| `outbox`, `processed_events` | every service | that service | that service's relay and consumers |
| Keycloak tables | keycloak | Keycloak | Keycloak only |

One Postgres container, four databases, four roles, each role granted only its own database; `currency` has none. A service that tries to read another's tables fails at connection time, which is the cheapest way to keep the boundary honest.

## 9. Justifications

| Decision | Why | Rejected |
|---|---|---|
| Nginx at the edge | TLS, per-IP limits and static files without Go code. Universally documented. | Caddy: simpler certificates, fewer knobs. Envoy: stronger, heavier. A separate API gateway product: nothing for it to do that the BFF does not. |
| BFF verifies tokens | Open-source Nginx cannot; every service does anyway. | Nginx Plus, njs, OpenResty: paid or extra runtimes. |
| Four services, not one | The goal is to learn service boundaries, events and ownership. The boundaries follow the PRD's own sections: §7 rows, §8 rules, §9 identity, §10 notifications, D35 rates. | Modular monolith: simplest, teaches least. Ten services: nothing left to own. |
| All of §8 inside `ledger` | §8.9's single transaction, and every rule reads every table. | A separate recap service: three network calls per number and a saga for currency change. |
| Replicated member facts | `ledger` and `notify` need timezone, `signed_up_on` and email on every computation or send. Events keep the copies fresh without a call per request. | Calling `members` per request: a dependency in the hot path. Sharing the database: not a boundary. |
| RabbitMQ | The textbook queue: routing, dead letters, quorum queues, per-message acks. MPL-2.0. Streams exist if replay is ever needed. | NATS: lightest, less to learn. Kafka: heaviest, replay not needed yet. Postgres queue: simplest, no broker skills. |
| Redis in phase 1 | Five real jobs (§4.4). AGPL binds Redis, not the app. | Valkey: identical API, drop-in if policy changes. No cache: fewer parts, no idempotency store. |
| Protobuf, Buf, Connect | One schema for RPCs and events; Go, Swift, TS and Kotlin generated; plain HTTP through Nginx. | OpenAPI: fine for phone-to-BFF, no event types. GraphQL: for many clients, not one. gRPC alone: needs HTTP/2 end to end and a proxy for browsers. |
| Outbox and `processed_events` | The only pattern under which "Spend saved but event lost" and "event delivered twice" are both impossible to observe. | Publish after commit without outbox: loses events on a crash. Exactly-once from the broker: does not exist end to end. |
| Server-side computation | One implementation of §8, one answer key, no drift. | Phone-side arithmetic: a second implementation in Swift. |
| Synchronous `EvaluateWarning` | The banner is in the Save response, as §8.8's "after saving a Spend" reads. | Event-only: the banner appears on the next `GetHome`, a second later. |
| Idempotency enforced in the owning service | A unique key inside the write's own transaction survives a BFF crash and a Redis restart; a cache cannot. | Redis as the guarantee: a crash between commit and cache write duplicates the Spend. |
| Originals kept on conversion | Two columns preserve what the Member typed; the rewrite the PRD asks for still happens. | Convert on read: reverses D12 and puts rounding in every read. Rewrite only: the typed amount is gone. |
| `next_*_at` columns for schedules | An indexed range scan instead of `AT TIME ZONE` per row per minute; batched RPCs and a worker pool for the fan-out. | Per-row local-time query: a full scan every tick and one call per Member at a popular time. |
| `notify` optional on every Member read and write path | A notification service must never block a money entry or blank Home; the event path already guarantees the flag and the push. The two paths that need it are `RequestMagicLink`, sent before sign-in, whose whole purpose is that `notify` emails the link, and `RequestExport`, whose whole purpose is that `notify` emails the CSV. | Synchronous, required: `notify` down means no Spends can be logged. |
| Login Token, not the action-token magic link | The action token lives on Keycloak's host and is consumed by any GET; the login token is opaque, app-redeemed, atomically single-use. | Action-token link with `send_email:false` and our own wrapper page: workable, but a scanner that follows the unwrapped URL still burns it. |
| SOPS + age as the one store, decrypted on the runner | Secrets encrypted in git, no daemon, MPL-2.0 and BSD-3. The deploy decrypts on the GitHub runner and sends the machine a mode-0600 `.env` and Keycloak's two vault files; the age key never reaches the machine (§10 "Secrets (A13)"; owner, 2026-09-22, PR #85). | Vault/OpenBao: a server for one VM. Doppler/Infisical: a hosted service for one developer. A plain `.env` as the store: nothing encrypted to commit. Files only, decrypted under `/run/secrets` (spec #61, until 2026-09-22): Keycloak cannot read its own database or admin password from a file, and anyone able to run `docker inspect` on the machine can already read any file there. |
| Apple identity-provider extension, not generic OIDC | It signs the client secret from the `.p8` on every exchange, so the six-month rotation job does not exist. | Generic OIDC with a hand-rotated JWT: a recurring operations task that fails silently when forgotten. |

## 10. Containers in phase 1

| Container | Image | Ports | Depends on |
|---|---|---|---|
| nginx | nginx:stable | 443, 80 | bff, keycloak |
| acme | neilpang/acme.sh | none | shares the cert volume with nginx; deploy hook reloads nginx |
| bff | built from the one Go module `apps/backend`, command `cmd/bff` (the one `Dockerfile`, `CMD=bff`) | 8080 internal | `ledger`, which waits for `members`, and redis; it calls `notify` and `currency` without waiting on them (A14, #175) |
| members, ledger, notify, currency | built from the one Go module `apps/backend`, command `cmd/<service>` (the one `Dockerfile`, `CMD=<service>`) | 8080 internal each | postgres and rabbitmq for `members`, `ledger` and `notify` (A4, A5); `currency` only redis |
| keycloak | quay.io/phasetwo/phasetwo-keycloak, pinned tag (§7.5) | 8080 internal | postgres |
| postgres | postgres:17 | 5432 internal | volume |
| rabbitmq | rabbitmq:4-management | 5672, 15672 internal | volume |
| redis | redis:8 | 6379 internal | none |
| otel-lgtm | grafana/otel-lgtm | 4317 internal, 3000 admin-only | volume |
| pg-backup | built from `infra/pg-backup` (cron + `pg_dump`) | none | postgres |

Thirteen containers, one `compose.yaml`, a health check on every long-running one but the `acme` sidecar, so `depends_on: condition: service_healthy` starts them in order. `otel-lgtm` and `pg-backup` are planned; `infra/compose/README.md` lists what `compose.yaml` runs today. Migrations run as a one-shot step per service before that service starts: the same image with the `migrate` subcommand, which `depends_on: condition: service_completed_successfully` orders (`members-migrate` is the first, #15). The same compose file runs on a laptop and on the first VM. Kubernetes comes after phase 1 is stable.

**The thirteen are the VM's.** A laptop runs the same file plus Mailpit, the one container that never reaches the VM: behind the `mail` profile, it is the SMTP host `notify` sends to (#68). The VM's `notify` has no mail host until #87 item 8. `realm-apply`, a one-shot step outside the count like the migration steps, builds the realm on every `up` (#65), and the deploy runs it on the machine too (#73, PR #85). Development-only containers are outside the count on any machine. Nginx and the `acme.sh` sidecar are two of the thirteen already, so joining the laptop stack (#72) changes no number: what changes is that there they serve plain HTTP and the sidecar holds no certificate, because ACME needs a domain.

**What this does and does not buy.** The service split gives independent deployment and clean ownership. It does not give independent scaling: one VM, one Postgres process and one disk are shared, so a runaway query in `ledger` slows Keycloak and `notify` too. Nothing in phase 1 should be designed as if the four services could be scaled apart; that arrives with separate database instances, after phase 1.

**Secrets (A13).** Every secret is to live in `infra/secrets/prod.yaml`, encrypted with SOPS and an age key, and committed once it holds real values (#87, `infra/secrets/README.md`). The deploy decrypts it on the GitHub runner, never on the VM, and sends the VM a `.env` (mode 0600) and Keycloak's two identity-provider secrets as files. `members`, `notify` and `ledger` read their connection strings, and the broker's URL, from files Compose mounts with `secrets:` (`*_FILE`), Keycloak reads the two provider secrets through its file vault, and Postgres, Keycloak's own passwords, and the BFF, `members` and `notify` their client secrets, are environment variables, exactly as on a laptop: Keycloak cannot read its own passwords from a file, and anyone able to run `docker inspect` on the VM can already read any file there (`infra/README.md`, "Deploying"). The age private key in the GitHub `production` environment is the only value a human ever holds; the VM never has it. Provider credentials rotate by edit, commit, deploy; the stack's own passwords are chosen once, when their volume is first initialised. Vault and OpenBao were rejected as a server for one VM; Doppler and Infisical as a hosted service for one developer.

**Infrastructure as code.** `infra/tofu/` holds the Keycloak realm, one root the stack applies on every `up` (`infra/tofu/realm/`, #65). A second root with its own state is to hold the DNS records, the VM with cloud-init and firewall, and API tokens for providers that expose them (Resend or Postmark for SMTP); it is #87's, and is not written until the hosting company is chosen there. The apple-app-site-association file is not OpenTofu's: Nginx serves it as a static file (`infra/nginx/html/`, #72). What stays human is one browser session each: Apple Developer Program enrolment, the App Store Connect API key, Apple's two `.p8` keys and Services ID, the Google OAuth client, the Open Exchange Rates App ID, the first SMTP-provider account, the domain purchase, the cloud account, and one `age-keygen`. The setup wizard is scoped to exactly those, each ending in pasting the value into `sops edit infra/secrets/prod.yaml`.

**Delivery (A15).** One long-lived branch, `main`. CI (`.github/workflows/ci.yml`) runs on every pull request and on `main`. `quality` runs the pre-commit hooks, skipping the Go, Swift, Buf and OpenTofu ones that run in their own jobs below and gitleaks, which `security.yml` runs over the whole tree, and always fires. Four jobs then run only when the files they build or check changed, generated output included, decided by `dorny/paths-filter` in a `changes` job: `design-system` (generated tokens, Swift tokens and inlined artwork are fresh; typecheck; TypeScript coverage), `contract` (`buf lint`, `buf breaking` against `main`, generated code is fresh), `backend` (golangci-lint, `tofu fmt` and `tofu validate` on every OpenTofu root, the secrets decrypt and wizard check smoke tests on pinned `sops` and `age`, the Compose stack with Mailpit, then `go vet`, every Go test including the Connect-boundary, end-to-end, magic-link and edge tests, Go coverage, `infra/compose/a10-smoke.sh` on the pinned Keycloak image, and `infra/compose/realm-noop.sh`) and `ios` (SwiftFormat, SwiftLint, build and unit tests on a simulator picked at runtime, iOS coverage). A final `ci` job gathers their results and is the one to require, because a job skipped by `if:` reports success while a workflow skipped by a `paths:` filter reports nothing at all and would block a pull request forever. The gating keeps a change's wait short: a change outside the app never starts the macOS job. `.github/workflows/security.yml` runs alongside it (govulncheck, OSV-Scanner, Semgrep, Trivy and a full-tree gitleaks scan, gathered by a final `security` job) on every pull request, every push to `main` and every Monday. Its jobs are not gated: all of them run on Linux, and govulncheck reports only the advisories the Go code reaches, so its verdict can change while `go.mod` does not. Merging to `main` means tested and integrated, not deployed. Deployment is a tag: pushing `vX.Y.Z` runs a deploy workflow in the GitHub `production` environment that first decrypts `infra/secrets/prod.yaml` with the environment's age key, then builds the images and pushes them to GHCR, copies the Compose file and the edge configuration, and runs Compose on the VM (`.github/workflows/deploy.yml`, `infra/README.md` "Deploying"); the iOS build for the same tag goes to TestFlight, a half `deploy.yml` does not have yet, because it needs the App Store Connect API key (spec #61). No `development` branch: for one developer merging one ticket at a time, git-flow adds a second merge per change and lets finished work sit unreleased. A ruleset on `main` is the server-side gate: no deletion, no force push, every change through a pull request with linear history, and `ci` and `security` green before merge, with an empty bypass list. A tag ruleset lets only the repository admin create, move or delete a `v*` tag, and the `production` environment deploys only from a `v*` tag after the admin approves.

**Backups (A12).** `pg-backup` runs `pg_dump` of each of the four databases at 03:00 UTC to an S3-compatible bucket in the same region, encrypted with a key from the secrets file, and keeps 30 daily copies. RabbitMQ and Redis hold nothing that cannot be rebuilt: an unpublished outbox row is re-published by the relay, and every Redis key is a cache or a short-lived lock. Once a quarter, a restore into a scratch Postgres is run and the ledger fixtures are executed against it; a backup that has never been restored is not a backup.

## 11. Decisions

| # | Decision | Status |
|---|---|---|
| A1 | Contract: protobuf, Buf, Connect; `buf breaking` required in CI; event messages in the same package. | in use: every RPC of §6 and §7, the screen views and every event of §4.3 are generated from `packages/contract`; `buf breaking` at `PACKAGE` level |
| A2 | All of §8 lives in `ledger` and nowhere else, including the warning-line check and conversion rounding; the phone stores and formats; the §8 fixtures are the answer key. One exception: onboarding step 4's preview is the Salary times the slider's percent, worked out on the phone as the knob moves, since a round trip per step of 5 would lag the drag; it is never stored, sent or read back, and the saved Target's amount is `ledger`'s. | in use: `ledger` computes every Home figure, every Recap, the Streak run, the Archetype title and the currency change (§8.9); every fixture passes through it |
| A3 | `member.timezone` as an IANA name; every "today" in §8 and §10 evaluated in it; the `notify` scheduler is the §4.5 query. | in use: the app sends `TimeZone.current.identifier` on every launch; `members.SetTimezone` checks it against Go's embedded zone data (`time/tzdata`, linked in because the runtime image has no zone files) and stamps `signed_up_on` once; `ledger` and the BFF read "today" in it; `notify`'s two schedules are the §4.5 query, each fire time computed in the zone |
| A4 | Events: outbox relay, at-least-once, `processed_events` dedupe keyed by message id. | in use: `members`, `ledger` and `notify` each run an outbox, relay, consumers and `processed_events` through `internal/events`; the other events are planned |
| A5 | Broker: RabbitMQ, topic exchange, quorum queues, dead-letter after five tries. | in use: `mmr.events`, `mmr.dlx` and the quorum queues `ledger.members`, `notify.ledger` and `notify.members`, each with its `.dead` queue, five tries (§4.3) |
| A6 | Redis for rate limits, idempotency, rate cache and JWKS; fail-open and fail-closed rules per job; AGPL noted. | in use: the magic-link limit (failing closed), the per-Member limit of 120 per minute (failing open), the idempotency fast path (failing open) and `currency`'s day's rate table. Not built: the JWKS cache |
| A7 | Front door: Nginx for TLS, per-IP limits and static files; the BFF verifies the JWT; every service verifies again. | in use on the laptop: the four routes, the per-IP limits and the two static files, over plain HTTP; the BFF and each service verify the JWT, the two pre-authentication procedures excepted by name (§4.2). Not live: TLS, since there is no domain yet |
| A8 | Service boundaries: `members`, `ledger`, `notify`, `currency`; payment and AI reserved; replicated member facts fed by events. | in use: `members`, `ledger` and `notify` each with its own database, `currency` with none; `ledger` and `notify` keep member-fact replicas fed by events (A4), with a read on a miss as the safety net |
| A9 | Data: one Postgres container, four databases, one role each, no cross-database grants. Deployment: Docker Compose for phase 1. | in use: four databases, `currency` holding none |
| A10 | Email sign-in: Phase Two Login Token redeemed by the app through the direct grant; the BFF requests it and limits 5 per email per hour; `notify` emails the app-domain link; the static fallback never touches the token; the Phase Two image is pinned. | in use on tag 26.6.6.1789466980 (`infra/compose/a10-smoke.sh`): the BFF requests the login token and limits five per address per hour, `notify` emails the link, the app redeems it, Nginx serves the `/magic` fallback. Every link so far goes to Mailpit: there is no SMTP provider and no domain yet |
| A11 | Observability: OpenTelemetry in every Go process, OTLP to one `otel-lgtm` container, trace context carried in Connect and RabbitMQ headers. | proposed |
| A12 | Backups: nightly `pg_dump` per database to object storage, 30 days kept, quarterly restore drill. | proposed |
| A13 | Secrets: SOPS + age, encrypted in git, decrypted on the deploy runner; the machine gets a `.env` and Keycloak's two vault files, never the age key; Keycloak file vault; OpenTofu for DNS, VM, realm and provider tokens; the wizard covers only the nine human steps. | in use in part: SOPS + age and secrets as files, the OpenTofu realm, the wizard, the Keycloak file vault (Google's secret; Apple's `.p8` not yet, §7.5), and the deploy's `.env` and vault files (`infra/compose/machine-env.sh`), written and never run. Not done: the age key, DNS and the VM |
| A14 | Every write RPC carries an idempotency key; the owning service enforces it with a unique index inside the write's transaction; Redis is only a fast path. `notify` is optional on every Member path; `RequestMagicLink`, sent before sign-in to have `notify` email the link, and `RequestExport`, to have it email the CSV, need it. Schedulers insert `sent_log` before sending. Conversions keep the typed originals. | in use in part: every write in §7 carries `idempotency_key`; `members`, `ledger` and `notify` claim it in the write's transaction through `internal/idempotency` (a repeat replays, a reused key for a different request is `invalid_argument`, keys expire after 24 hours); the BFF refuses a Member write without the header and answers retries from Redis first (§4.4); the Target Warning push claims its `sent_log` row before sending. Not built: the remaining Member writes and the conversions |
| A15 | Trunk-based delivery: `main` is the only long-lived branch; every change is a short branch and a pull request with CI green; deployment is a separate step triggered by a version tag, run by a GitHub Actions job in the `production` environment; a staging server, if ever needed, is another environment, never a branch. | in use (`AGENTS.md` "Branching"); `.github/workflows/deploy.yml` is written and linted but has never run; its first tag push is #87's |
