# infra/nginx

The edge. Today it serves the page a sign-in link opens and the Apple app link file. Its `/api/` and `/auth/`
routes also work, but the app still calls the BFF and Keycloak on their own ports. One file, `default.conf`, holds
its routes and limits, and `html/` holds the two static files.

## Where a request goes

The two static files are the sign-in fallback page and the Apple app link file. Anything else gets 404.

```mermaid
flowchart TD
    phone["Phone or browser"]
    nginx["Nginx edge"]
    bff["BFF"]
    keycloak["Keycloak"]
    files["Two static files"]
    phone --> nginx
    nginx -- "/api/" --> bff
    nginx -- "/auth/" --> keycloak
    nginx -- "/magic, app link" --> files
```

## How it works

- **`/api/`** goes to the BFF with the prefix removed: `/api/money_recap.v1.BffService/GetHome` arrives as
  `/money_recap.v1.BffService/GetHome`.
- **`/auth/`** goes to Keycloak. On a laptop every caller still uses Keycloak's own port.
- **`/.well-known/apple-app-site-association`** is served as `application/json` with no redirect, as Apple
  requires. Its Team ID is still a placeholder: nothing writes the real one in yet.
- **`/magic`** is the page an email link opens on a device without the app. It has no script, so it never uses
  the token.
- **Limits per IP address**: 60 requests a second with a burst of 120, and 100 connections, on `/api/` and
  `/auth/`. The sign-in link call alone is held to 5 a minute.
- **The limit on sign-in links per email address** is the BFF's ([bff README](../../apps/backend/cmd/bff/README.md)),
  because Nginx cannot read an address inside a request.
- **TLS**: none yet. Nginx serves plain HTTP on port 80. The `acme` container shares a volume with it, ready
  for a certificate once there is a domain ([#87](https://github.com/ravialdyh/my-money-recap-monorepo/issues/87)).

## Test it

With the stack up ([compose README, "Run it"](../compose/README.md#run-it)):

```bash
curl -si http://localhost:8090/.well-known/apple-app-site-association | grep -i '^HTTP/\|^Content-Type'
```

```bash
curl -s http://localhost:8090/magic | grep -o 'Open My Money Recap[^<]*'
```

`TestEdge` in `apps/backend/internal/bff` checks the `/api/` route, both files and the sign-in link limit.
