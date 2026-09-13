# Setup automation: what is code, what is a human

| | |
|---|---|
| Date | 2026-09-14 |
| Question | Which of the "human-only setup" steps in `ARCHITECTURE.md` (§5, §7.5, §10) can be infrastructure-as-code, how secrets should be held for a one-VM Docker Compose deployment, whether the Apple client-secret rotation job is needed at all, and what genuinely stays human. |
| Method | Primary sources only: official docs, source repositories, provider docs. Every claim carries a URL. Items I could not confirm are marked **unverified**. |

## Summary

The single most important finding: **the Apple client-secret rotation problem does not exist for this stack.** The Apple identity provider that ships inside the Phase Two Keycloak container (`klausbetz/apple-identity-provider-keycloak`) takes the `.p8` key, Team ID and Key ID as configuration and signs a fresh client-secret JWT itself on every authorization-code exchange. Nothing expires, no cron, no script to "commit, not remember". Details in §3.

| Step in the plan | Automatable? | With what | What stays human |
|---|---|---|---|
| Domain | Partly | DNS records: `cloudflare_dns_record` (or any registrar's Terraform provider) | Buying the domain; creating the Cloudflare API token |
| Universal-links AASA file | Yes | A static file served by the edge proxy from `/.well-known/`; Apple's requirements are plain HTTPS + `application/json` + no redirect | — |
| Apple App ID + Sign in with Apple capability | Yes | App Store Connect API (Bundle IDs, Bundle ID Capabilities) or `fastlane produce --enable_services sign_in_with_apple` | — |
| Apple Services ID (web client id for Keycloak) | **No** | Not in the App Store Connect API; not in fastlane | Register in the developer portal once |
| Apple `.p8` keys (Sign in with Apple, APNs) | **No** | Not in the App Store Connect API | Create in Certificates, Identifiers & Profiles > Keys; download once |
| Apple client-secret JWT (6-month expiry) | **Eliminated** | The Keycloak Apple IdP extension signs it from the `.p8` at token time | — |
| Google OAuth client | **No** | No public API; the IAP API's clients are locked to IAP; `gcloud iam oauth-clients` is Workforce Identity Federation | Create in Google Auth Platform > Clients once |
| Keycloak realm, clients, IdPs, SMTP | Yes | Official `keycloak/keycloak` Terraform/OpenTofu provider (v5.9.0, 2026-07-31, Apache-2.0); secrets via Keycloak's file vault `${vault.x}` | — |
| Open Exchange Rates App ID | **No** | Dashboard only | Sign up, copy the App ID once |
| SMTP provider credentials | Yes (after sign-up) | Resend `POST /api-keys`; Postmark Account API creates a server and returns its tokens; SES SMTP password is derived from an IAM access key | Creating the provider account and its first account-level token |
| Hosting VM | Yes | Official provider of the host, e.g. `hetznercloud/hcloud` (MPL-2.0) with cloud-init | Creating the cloud account and its first API token |
| TLS certificates | Yes | Caddy (built in), Traefik (ACME resolver), or `acme.sh` sidecar with an Nginx reload hook. The `certbot/certbot` image explicitly cannot install or reload | — |
| Runtime secrets | Yes | SOPS + age: encrypted files committed to git, decrypted at deploy into Compose `secrets:` files under `/run/secrets` | Generating the one age key and keeping it in two places |
| Developer Program, terms, first API keys | **No** | — | Enrol, pay, agree; create the App Store Connect API key (Account Holder/Admin, downloadable once) |

The wizard therefore shrinks to about eight copy-paste values, each obtained once from a web console, and one `age-keygen`. Everything else is `tofu apply` plus `sops`.

## 1. What can be infrastructure-as-code, and with what

### 1.1 Keycloak: Terraform provider vs realm JSON vs kcadm

**The provider is official now.** The repository lives at `github.com/keycloak/terraform-provider-keycloak`; the README documents migrating state from `mrparkers/keycloak` to `keycloak/keycloak` with `terraform state replace-provider`, and the licence is Apache-2.0 ([repo](https://github.com/keycloak/terraform-provider-keycloak)). Latest release v5.9.0, 2026-07-31, tested against Keycloak 26.0.8 to 26.7.3 ([releases](https://github.com/keycloak/terraform-provider-keycloak/releases)). Resources this project needs:

- `keycloak_realm` with an `smtp_server { host, from, auth { username, password } }` block ([realm.md](https://raw.githubusercontent.com/keycloak/terraform-provider-keycloak/main/docs/resources/realm.md)).
- `keycloak_oidc_google_identity_provider` with `client_id`, `client_secret` ([oidc_google_identity_provider.md](https://raw.githubusercontent.com/keycloak/terraform-provider-keycloak/main/docs/resources/oidc_google_identity_provider.md)).
- `keycloak_oidc_identity_provider` with `provider_id` ("should be used unless you have extended Keycloak and provided your own implementation") and `extra_config` ("can be used for custom oidc provider implementations") — this is how the Apple extension's provider is declared; `client_secret` accepts the `$${vault.ID}` form ([oidc_identity_provider.md](https://raw.githubusercontent.com/keycloak/terraform-provider-keycloak/main/docs/resources/oidc_identity_provider.md)).

**Realm JSON import is a bootstrap, not configuration management.** Keycloak's own docs: "If a realm already exists in the server, the import operation is skipped", and "To re-create realms you should explicitly run the `import` command prior to starting the server." Placeholders such as `${MY_REALM_NAME}` resolve from environment variables, with the caveat to "take care to ensure placeholders references do not inappropriately expose sensitive environment variable values" ([Keycloak: Importing and exporting realms](https://www.keycloak.org/server/importExport)). So a committed `realm.json` applies once; every later change (a new redirect URI, a new IdP) is either a manual click or a delete-and-reimport.

**kcadm is imperative.** The admin guide lists realm, client, identity-provider and other operations for `kcadm.sh` but gives no declarative or idempotent mode ([Server Administration Guide, Admin CLI](https://www.keycloak.org/docs/latest/server_admin/index.html#admin-cli)). It is a fine tool inside a script; it is not a state model.

**Alternative:** `adorsys/keycloak-config-cli` applies JSON/YAML idempotently through the Admin API, supports environment-variable substitution, Apache-2.0, tracks the latest four Keycloak releases ([repo](https://github.com/adorsys/keycloak-config-cli)). It is the right answer if the team wants no Terraform at all; the official provider is the right answer if Terraform/OpenTofu is already in use for DNS and the VM, which this document recommends.

**Secrets inside Keycloak.** Keycloak has a vault SPI with a file provider ("especially useful for Kubernetes/OpenShift secrets", enabled with `kc.sh build --vault=file` and `--vault-dir=/path`). The `${vault.secretname}` placeholder is accepted in exactly the three places this project needs: "Obtain the SMTP Mail server Password", "Obtain the OIDC identity providers Client Secret", and LDAP ([Keycloak: Using a vault](https://www.keycloak.org/server/vault)). Mount the Compose secrets directory as the vault dir and neither the Terraform state nor the realm holds a secret value.

**Terraform vs OpenTofu.** HashiCorp Terraform (and Vault) moved to the Business Source License 1.1 ([Vault LICENSE](https://raw.githubusercontent.com/hashicorp/vault/main/LICENSE)); OpenTofu is the Linux Foundation fork under MPL-2.0, "a drop-in replacement for Terraform with registry compatibility" ([OpenTofu intro](https://opentofu.org/docs/intro/)). The Keycloak provider README names both Terraform and OpenTofu. Use OpenTofu; the HCL is identical.

### 1.2 Google OAuth client

No public API creates an ordinary OAuth 2.0 client ID:

- Google's programmatic OAuth client page is IAP-only: "OAuth clients created by the API are locked for IAP usage only" and "The API does not operate on the OAuth clients that were created using the Google Cloud console" ([Programmatically creating OAuth clients for IAP](https://docs.cloud.google.com/iap/docs/programmatic-oauth-clients)).
- `gcloud alpha iam oauth-clients` belongs to Workforce Identity Federation ("OAuth application integration is a feature of Workforce Identity Federation") ([Workforce OAuth apps](https://docs.cloud.google.com/iam/docs/workforce-oauth-app), [gcloud reference](https://docs.cloud.google.com/sdk/gcloud/reference/alpha/iam/oauth-clients)).
- The Terraform feature request for a non-IAP client resource (#16452, opened 2023-11-06) is closed without a resource ([issue](https://github.com/hashicorp/terraform-provider-google/issues/16452)); the older #6074 is the same ask ([issue](https://github.com/hashicorp/terraform-provider-google/issues/6074)).
- Google's own instructions say to "Navigate to the Google Auth Platform Clients page" in the console; secrets "are hashed. This means you will only be able to view and download the full client secret once", and the console offers client-secret rotation (add new, migrate, disable old, delete old) ([Manage OAuth Clients](https://support.google.com/cloud/answer/15549257)).

Human, once. The client id and secret are then two SOPS values.

### 1.3 Apple Developer resources

**App Store Connect API.** Its Provisioning section covers exactly: Bundle IDs, Bundle ID Capabilities, Certificates, Devices, Profiles, Merchant ID, Pass type Ids ([API index, JSON](https://developer.apple.com/tutorials/data/documentation/appstoreconnectapi.json)). There is no endpoint for Services IDs and none for Keys (`.p8`). So the App ID and its Sign in with Apple capability are API-creatable; the Services ID that Keycloak uses as `client_id`, and the Sign in with Apple and APNs keys, are portal-only ([Register a Services ID](https://developer.apple.com/help/account/identifiers/register-a-services-id/), [Create a private key](https://developer.apple.com/help/account/manage-keys/create-a-private-key/) — Account Holder or Admin, "you will not be able to download it again").

**fastlane.** `produce` will "Create new iOS apps on both the Apple Developer Portal and App Store Connect" and its `enable_services` accepts `sign_in_with_apple: on`; it does not create Services IDs or keys ([produce](https://docs.fastlane.tools/actions/produce/)). `match` handles code-signing certificates and provisioning profiles only ([match](https://docs.fastlane.tools/actions/match/)); `sigh` "can create, renew, download and repair provisioning profiles" ([sigh](https://docs.fastlane.tools/actions/sigh/)). None touch Sign in with Apple keys, APNs keys or Services IDs.

**Creating the API key itself** is human: Account Holder or Admin, in App Store Connect > Users and Access > Integrations, private key downloadable once ([Creating API Keys](https://developer.apple.com/documentation/appstoreconnectapi/creating-api-keys-for-app-store-connect-api)).

### 1.4 DNS and the apple-app-site-association file

- DNS: `cloudflare_dns_record { zone_id, name, type, content, proxied, ttl }` in the official Cloudflare provider v5 ([dns_record.md](https://raw.githubusercontent.com/cloudflare/terraform-provider-cloudflare/main/docs/resources/dns_record.md)). Any registrar with a provider works the same way.
- AASA: Apple requires the file at `/.well-known/apple-app-site-association`, over HTTPS with a valid certificate, `Content-Type: application/json`, no redirects, reachable by Apple's CDN ([Supporting associated domains](https://developer.apple.com/documentation/xcode/supporting-associated-domains)). That is a committed static file behind the edge proxy — already what §5 says. If the proxied host is Cloudflare, keep the record `proxied = false` or confirm Cloudflare does not rewrite the response; **unverified** here. Phase Two hosted custom domains can also serve AASA ([Phase Two blog](https://phasetwo.io/blog/app-association-custom-domains/)), not relevant for self-hosting.

### 1.5 Let's Encrypt for the edge

| Option | What it does | Source |
|---|---|---|
| `certbot/certbot` container | Obtains certs only. Docs: "this mode of operation is unable to install certificates or configure your webserver", and "Most users should use the instructions at certbot.eff.org. You should only use Docker if you are sure you know what you are doing." | [certbot install docs](https://eff-certbot.readthedocs.io/en/stable/install.html#running-with-docker) |
| Caddy | "the first web server to use HTTPS automatically and by default"; obtains and renews from Let's Encrypt/ZeroSSL, redirects 80 to 443, "no downtime, extra configuration, or separate tooling is required". Licence Apache-2.0 per its repository (**not re-verified in this pass**). | [Caddy automatic HTTPS](https://caddyserver.com/docs/automatic-https) |
| Traefik | ACME certificate resolvers with `httpChallenge`, `tlsChallenge`, `dnsChallenge`; "starts renewing them 30 days before their expiry"; stores in `acme.json`; single-instance only for ACME. | [Traefik ACME](https://doc.traefik.io/traefik/reference/install-configuration/tls/certificate-resolvers/acme/) |
| acme.sh | "A pure Unix shell script ACME client", GPLv3, daily cron renewal with ARI (RFC 9773), an Nginx mode, deploy hooks that reload Nginx, DNS APIs for most providers, official Docker image. | [acme.sh](https://github.com/acmesh-official/acme.sh) |

Two of the thirteen containers in §10 exist only to hold a certificate (nginx + certbot), and the certbot one cannot do the reload the table promises. Caddy collapses them into one container with the same routes; if A7's Nginx choice stands, replace certbot with an `acme.sh` sidecar whose deploy hook runs `nginx -s reload`.

### 1.6 Open Exchange Rates, SMTP, hosting

- **Open Exchange Rates**: "If you need to create or deactivate an App ID, please visit your Account Dashboard." No API ([OER auth docs](https://docs.openexchangerates.org/reference/authentication)). Human, once.
- **Resend**: `POST https://api.resend.com/api-keys` with `name` and `permission` (`full_access` | `sending_access`) ([Create API key](https://resend.com/docs/api-reference/api-keys/create-api-key)). Terraform providers exist only from the community (`jhoward321/resend`, `RealIanX/resend`, `chronark/terraform-provider-resend`) ([registry search](https://registry.terraform.io/providers/jhoward321/resend/latest/docs)); none is official.
- **Postmark**: `POST /servers` with `X-Postmark-Account-Token` creates a server and the response includes `ApiTokens` ([Servers API](https://postmarkapp.com/developer/api/servers-api)). Community Terraform providers only (`marcus-carey/postmark`, `shebang-labs/postmark`).
- **Amazon SES**: the SMTP password is derived from an IAM secret access key by a documented HMAC algorithm ("the SMTP user name is the same as the AWS access key ID, so you only need to generate the SMTP password") ([SES SMTP credentials](https://docs.aws.amazon.com/ses/latest/dg/smtp-credentials.html)). This is the only SMTP option whose credential is fully Terraform-managed with an official provider (IAM user + access key, then the derivation). The exact attribute name on `aws_iam_access_key` (`ses_smtp_password_v4`) is from memory — **unverified** here.
- **VM**: `hetznercloud/terraform-provider-hcloud` is the official Hetzner provider, MPL-2.0 ([repo](https://github.com/hetznercloud/terraform-provider-hcloud)); cloud-init `user_data` and firewall resources are documented in its registry docs (**not fetched**). Equivalent providers exist for DigitalOcean, AWS, GCP.

## 2. Secrets for one VM, one developer, Docker Compose

What Docker itself says: "If you're injecting passwords and API keys as environment variables, you risk unintentional information exposure"; Compose `secrets:` are "mounted as a file in `/run/secrets/<secret_name>` inside the container" and "Compose grants access to secrets on a per-service basis" ([Compose: secrets](https://docs.docker.com/compose/how-tos/use-secrets/)). So the target is files under `/run/secrets`, not a `.env`.

| Tool | Open source / licence | Daemon needed with Compose? | Rotation | Notes and source |
|---|---|---|---|---|
| SOPS + age | SOPS MPL-2.0, CNCF sandbox since 2023 ([repo](https://github.com/getsops/sops)); age BSD-3-Clause ([repo](https://github.com/FiloSottile/age)) | No. A binary and a key file. `sops exec-env` / `sops exec-file` pass plaintext to a subprocess "only passed in memory ... the plaintext secrets never touch the disk" ([advanced usage](https://getsops.io/docs/usage/advanced/)) | `sops rotate` "generates a new data encryption key and reencrypt all values"; `sops updatekeys` adds/removes recipients per `.sops.yaml` ([key management](https://getsops.io/docs/usage/key-management/)). Rotating a provider secret is: edit the encrypted file, commit, deploy | Recipients in `.sops.yaml` `creation_rules`; key from `SOPS_AGE_KEY_FILE` ([age docs](https://getsops.io/docs/usage/identities/age/)). Encrypted files live in git; history is the audit trail |
| Infisical | Core MIT, `ee/` proprietary ([repo](https://github.com/Infisical/infisical)) | Yes: a server (self-hosted = more containers, Postgres, Redis) or Infisical Cloud. `infisical run` injects env | Built-in rotation for some databases; dynamic secrets | Good product, wrong size: a control plane larger than the app it protects |
| Doppler | Platform is a hosted service; only the CLI is Apache-2.0 ([cli repo](https://github.com/DopplerHQ/cli)) | No daemon, but a network dependency at every `doppler run -- docker-compose up`; production needs a Service Token ([Docker Compose guide](https://docs.doppler.com/docs/docker-compose)) | In the dashboard/API | Vendor lock-in for a solo project |
| 1Password CLI / Connect | Proprietary | No daemon for `op run --env-file` ("runs the provided command in a subprocess with the secrets made available as environment variables only for the duration of the process"); service accounts for servers ([1Password CLI](https://www.1password.dev/cli/secrets-environment-variables/)) | In the vault | Fine if the developer already pays for 1Password; env-var injection, not files |
| HashiCorp Vault | BSL 1.1 since 1.15 ([LICENSE](https://raw.githubusercontent.com/hashicorp/vault/main/LICENSE)) | Yes: a server with unseal, storage, policies | Best-in-class dynamic secrets and leases | Sized for a platform team |
| OpenBao | MPL-2.0, Linux Foundation fork of Vault ([repo](https://github.com/openbao/openbao)) | Yes: `bao server` | Same as Vault | Same size problem |
| GitHub Actions environment secrets | Hosted, proprietary | N/A — CI only. "These secrets are only available to workflow jobs that use the environment" and are gated by protection rules ([environments](https://docs.github.com/en/actions/how-tos/deploy/configure-and-manage-deployments/manage-environments)); 48 KB and 100 per environment ([limits](https://docs.github.com/en/actions/reference/security/secrets)); not passed to forks ([use secrets](https://docs.github.com/en/actions/how-tos/write-workflows/choose-what-workflows-do/use-secrets)) | Overwrite in the UI/API | Right for the deploy pipeline's few inputs (SSH key, age key); wrong as the runtime secret store |

**Standard recommendation at this scale: SOPS + age.** It is the pattern the CNCF adopted, it needs no server, the encrypted files are reviewed and versioned like any other code, and the whole trust root is one age private key. Concretely:

1. `infra/secrets/prod.env.sops` (or one file per secret) encrypted to the age public key in `.sops.yaml`.
2. `deploy.sh`: `sops decrypt` each value into `/opt/mmr/secrets/<name>` (mode 0400), then `docker compose up`. Compose maps them as `secrets:`; Go services read `_FILE`-style paths; Keycloak mounts the same directory as `--vault-dir` and the realm refers to `${vault.prod_apple_p8}` etc.
3. The age private key exists in two places: the VM (`/root/.config/sops/age/keys.txt`) and a GitHub environment secret for the deploy workflow. It is the only secret a human ever handles.
4. Rotation of any provider secret = edit the encrypted file, commit, deploy. Rotation of the age key = `age-keygen`, `sops updatekeys`, replace the two copies.

## 3. The Apple client-secret JWT: rotation is not needed

**Apple's rule.** "It's an error to request an expiration time more than `15777000` seconds (six months) in the future, as measured by the clock on Apple's servers." Claims: `iss` = 10-character Team ID, `iat`, `exp`, `aud` = `https://appleid.apple.com`, `sub` = the App ID or Services ID used as `client_id`; signed ES256 ([Creating a client secret](https://developer.apple.com/documentation/accountorganizationaldatasharing/creating-a-client-secret)).

**Keycloak core has no Apple provider and will not add one.** The out-of-the-box social providers are Bitbucket, Facebook, GitHub, GitLab, Google, Instagram, LinkedIn, Microsoft, OpenShift 4, PayPal, Stack Overflow, Twitter — no Apple ([Server Administration Guide, identity broker](https://www.keycloak.org/docs/latest/server_admin/index.html#_identity_broker)). Issue #12468 "Add Apple ID to Identity Providers", which explicitly asked for "automated client secret renewal using Apple's private key", is closed as not planned ([issue](https://github.com/keycloak/keycloak/issues/12468)).

**The extension derives the secret from the key.** `klausbetz/apple-identity-provider-keycloak` (Apache-2.0) is configured with Client ID (Services ID), Team ID, Key ID and the raw `.p8` content: "The `/token`-request must contain a client-secret (JWT) which is signed using a specific private key (`.p8` file)" ([README](https://raw.githubusercontent.com/klausbetz/apple-identity-provider-keycloak/main/README.md)). In the source, `exchangeAuthorizationCode()` calls `prepareClientSecret()` → `generateClientToken()`, which builds the JWT with `iss` = Team ID, `sub` = client id, `aud` = `https://appleid.apple.com`, `exp = iat + 86400 * 180`, and signs it ES256 with the P8 key — on every code exchange ([AppleIdentityProvider.java](https://github.com/klausbetz/apple-identity-provider-keycloak/blob/main/src/main/java/at/klausbetz/provider/AppleIdentityProvider.java)). Latest release 1.17.0 (2025-01-15) targets Keycloak >= 26.5.0 ([releases](https://github.com/klausbetz/apple-identity-provider-keycloak/releases)). The older `BenjaminFavre/keycloak-apple-social-identity-provider` does the same ("Fill Client secret with the base 64 content of your private key file", plus Team ID and Key ID) ([README](https://github.com/BenjaminFavre/keycloak-apple-social-identity-provider)).

**It is already in the container this project plans to run.** The Phase Two Keycloak image bundles, alongside Magic Link, an "Apple Identity Provider — Enables Sign in with Apple for web-based and native applications (via token-exchange)" ([phasetwo-containers](https://github.com/p2-inc/phasetwo-containers)). Whether the bundled version tracks the Keycloak version in the image is the same open item §7.5 already has for Magic Link; **unverified** here.

**Consequence.** The `.p8` is a static secret (SOPS value → Keycloak file vault → `${vault.prod_apple_p8}` in the IdP's `client_secret`). There is no six-month expiry, no rotation script, and §7.5 "Operations" row goes away. The Terraform side is one `keycloak_oidc_identity_provider` with `provider_id = "apple"` and Team ID / Key ID in `extra_config` (exact config key names come from the extension's `AppleIdentityProviderConfig`; **not fetched** here).

**If a rotation job were ever needed** (a different IdP implementation): the established shape is a scheduled job that re-signs the JWT from the `.p8` and pushes it through the provider's management API. The open-source precedent is the Supabase `apple-client-secret-rotator` GitHub Action (scheduled workflow, zero dependencies, updates the project via the Management API) ([blog](https://blog.mansueli.com/automating-sign-in-with-apple-secret-rotation-on-supabase) — the page returned 403 to this fetch; description taken from the search snippet, **unverified**); generator scripts such as [Faruqt/generate-apple-client-secret-key](https://github.com/Faruqt/generate-apple-client-secret-key). For Keycloak the equivalent would be `tofu apply` from cron with the JWT as a variable, or `kcadm update identity-provider/instances/apple`. Not needed here.

## 4. What stays human

Each item is done once, in a browser, and produces at most one value for SOPS:

1. **Apple Developer Program enrolment**: Apple Account with two-factor auth, identity verification, "agreed to the associated program license agreement", USD 99/year ([Program enrollment](https://developer.apple.com/help/account/membership/program-enrollment/)). Agreements in App Store Connect are re-accepted by the Account Holder when Apple updates them.
2. **App Store Connect API key**: Account Holder/Admin, Users and Access > Integrations, downloadable once ([doc](https://developer.apple.com/documentation/appstoreconnectapi/creating-api-keys-for-app-store-connect-api)). This unlocks everything in §1.3 that is API-driven.
3. **Apple Services ID** for Keycloak's `client_id`, with the domain and return URL, and **two `.p8` keys** (Sign in with Apple; APNs), each downloadable once ([Services ID](https://developer.apple.com/help/account/identifiers/register-a-services-id/), [Keys](https://developer.apple.com/help/account/manage-keys/create-a-private-key/)).
4. **Google OAuth client** in Google Auth Platform > Clients ([doc](https://support.google.com/cloud/answer/15549257)); consent-screen brand fields too.
5. **Open Exchange Rates account** and App ID ([doc](https://docs.openexchangerates.org/reference/authentication)).
6. **SMTP provider account** and its first account-level token (Resend, Postmark) or an AWS account (SES); after that, tokens are API/Terraform.
7. **Domain purchase** and the registrar/DNS API token (Cloudflare token scoped to one zone).
8. **Cloud account** for the VM and its API token.
9. **One `age-keygen`** and placing the private key on the VM and in the GitHub environment.

Everything downstream of these nine values is `tofu apply` and `docker compose up`.

## Recommended changes to ARCHITECTURE.md

1. **§7.5 "Configuration as code"**: replace "Realm export JSON committed to `infra/keycloak/realm.json` and imported on start" with the official `keycloak/keycloak` OpenTofu provider (realm, Google IdP, Apple IdP via `keycloak_oidc_identity_provider` + `extra_config`, the app's public client, SMTP block). Keep a JSON import only if a first-boot realm is needed before the provider can log in; note Keycloak skips import when the realm exists, so JSON cannot carry later changes.
2. **§7.5 "Operations"**: delete the six-month Apple client-secret row and the committed generator script. State that the bundled Apple IdP extension signs the client secret from the `.p8` on every token exchange, so the `.p8` is a static secret with no expiry.
3. **§7.5 secrets**: replace "come from environment" with Keycloak's file vault: `--vault=file --vault-dir=/run/secrets`, and `${vault.<realm>_<name>}` for the SMTP password and both IdP client secrets.
4. **§10 keycloak image**: name `quay.io/phasetwo/phasetwo-keycloak` explicitly (it bundles Magic Link and the Apple IdP) instead of "quay.io/keycloak/keycloak with the Phase Two provider"; fold the A10 open item into one check: the extension versions in the chosen image tag against its Keycloak version.
5. **§5 Certificates and §10 certbot row**: drop `certbot/certbot` (its docs say it cannot install or reload). Either swap Nginx for Caddy (one container, TLS built in, same three routes and the static AASA file — revisit the A7 "Rejected: Caddy" line) or keep Nginx and add an `acme.sh` sidecar with an `nginx -s reload` deploy hook. Twelve containers either way.
6. **New §10.x "Secrets"** (or a new A13): SOPS + age; encrypted files in `infra/secrets/`; decrypted at deploy into Compose `secrets:` files under `/run/secrets`; Go services read `*_FILE` paths; the age private key on the VM and in the GitHub `production` environment is the only hand-held secret. Rotation = edit, commit, deploy.
7. **New `infra/` layout**: `infra/tofu/` (DNS records, VM + cloud-init, firewall, Keycloak realm, SMTP-provider tokens where the provider has one), `infra/compose/`, `infra/secrets/`, `infra/nginx/` or `infra/caddy/`. `infra/README.md` line "Secrets never live here; they come from the environment" becomes "Secrets live here encrypted with SOPS; they are mounted as files."
8. **Wizard scope**: rename the "human-only setup" list to the nine items in §4 above, each ending in "paste the value into `sops edit infra/secrets/prod.yaml`". Remove universal-links hosting, Let's Encrypt, the Keycloak realm, the Apple client-secret JWT and SMTP credentials from the human list.
9. **§9 Justifications**: add a row "Secrets: SOPS + age. Rejected: Vault/OpenBao (a server for one VM), Doppler/Infisical (a service for one developer), `.env` files (Docker's own docs warn against env-var secrets)."
10. **§11**: add A13 (secrets) and amend A7 if Caddy is adopted; A10 note that the Apple IdP is in the same image and shares the version check.
