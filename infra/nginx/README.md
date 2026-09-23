# infra/nginx

The edge (`docs/ARCHITECTURE.md` §5, A7). One configuration file, `default.conf`, carries all four
routes §5 names, the two static files under `html/`, and §5's Limits row. Compose mounts the file at
`/etc/nginx/conf.d/default.conf`, the path the stock `nginx:stable` image already includes from
inside its `http` block, which is why `limit_req_zone` and `limit_conn_zone` can be declared here
rather than in a second file.

## The four routes

| Route | What it does |
| --- | --- |
| `/api/` | `proxy_pass` to `bff:8080` with the prefix stripped, so `/api/money_recap.v1.BffService/GetHome` reaches the BFF as `/money_recap.v1.BffService/GetHome`. Connect speaks plain HTTP/1.1, so an ordinary `proxy_pass` is the whole of it (§4.1). |
| `/auth/` | The same to `keycloak:8080`. On a laptop Keycloak also keeps its own published port and every caller uses it, so this route exists and nothing depends on it yet (#72). |
| `/.well-known/apple-app-site-association` | The static file, served from its exact location with `default_type application/json`. |
| `/magic` | The static fallback page. |

Anything else is `404`. Without that last location the image's own welcome page answers `/`.

## The two static files

`html/apple-app-site-association` has no extension, so nginx finds no type for it in `mime.types`
and falls back to the location's `default_type`. That is how the header comes out as exactly
`application/json`, which Apple requires, with no charset appended: the stock `nginx.conf` sets no
`charset` directive. It is served from a `location =` with `try_files`, so there is no directory
index and no trailing-slash redirect; Apple refuses a redirected association file
(`docs/research/2026-09-14-setup-automation.md` §1.4). Its `appIDs` entry is
`TEAMID.com.mymoneyrecap.app`: the bundle identifier is real, the team identifier is a placeholder.
The real one is `apple_team_id` in the secrets file, which the wizard's step 4 claims since #71, but
nothing writes it into this file yet: the deploy copies `html/` as it is.

`html/magic.html` is the A3 web fallback page of `docs/SCREEN-STATES.md`: violet ground, an
envelope, one sentence, no control. It is one self-contained file with no script and no request of
its own, so tapping the link on a laptop cannot spend the login token: the token is only ever
consumed by the app's own POST to the token endpoint (PRD §11, A10). `/magic` and `/magic?lt=…`
answer the same bytes, which `TestEdge` asserts. The envelope is drawn inline rather than served
from `my-money-recap/assets/illustrations/`, because the deploy copies only what `compose.yaml`
mounts ([`../README.md`](../README.md), "Deploying", step 5), and `my-money-recap/` is not among it.

## The limits

`limit_req_zone` keys on `$binary_remote_addr`, an IP address, and there are two zones.

`mmr_api` is 60 r/s with `burst=120 nodelay`, deliberately generous because a mobile carrier can put
thousands of phones behind one address. With `limit_conn` at 100 per address it is applied to the
`/api/` and `/auth/` routes, and not to the two static files.

`mmr_magic` is the stricter 5 r/min, and it is applied in one `location =` on `RequestMagicLink`'s
own procedure path under `/api/`. `burst=4 nodelay` is what makes "5 r/min" mean five at once and
then one per twelve seconds, rather than one per twelve seconds from the first request. Both zones
are named in that location, because nginx inherits `limit_req` only into a level that declares none
of its own.

The location is exact, never a prefix. A prefix would put the whole Connect API behind 5 r/min and
throttle the end-to-end suite; `TestEdge` proves the difference by sending six rapid requests to
each path and asserting that the edge refuses one of the magic-link six with its own `503` and none
of the `GetHome` six.

This zone is **not** PRD §11's five per email address per hour. `limit_req` cannot see an address
inside a request body, so the BFF enforces that limit itself in Redis, fail closed (§4.4, #69).

## TLS and the `acme.sh` sidecar

Nginx serves plain HTTP on port 80, published as `NGINX_HOST_PORT` (80 itself on the machine), and
holds no certificate on any stack yet. The `acme` container runs the image's `daemon` command, which is supercronic over a
crontab that calls `acme.sh --cron`; with no domain it has nothing registered and renews nothing.
Both containers mount the `acme-certs` volume, nginx read-only, so the wiring §10 describes exists
before there is anything to put in it. Issuing a certificate, the `listen 443` server and the deploy
hook that runs `nginx -s reload` all need a public name, so they
belong to #87, which also names the machine.
