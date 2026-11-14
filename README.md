# My Money Recap

A personal money tracker for adults, styled as a playful kids' app, that turns one month of
income, spending and saving into a Wrapped-style monthly recap.

Work in progress: slices 1–7 are done; the real deploy and the rest of phase 1 are planned in #8 and #87.

## Layout

| Path | What lives here |
| --- | --- |
| `docs/` | The product rules (`PRD.md`), every screen state (`SCREEN-STATES.md`), the glossary (`CONTEXT.md`), and the backend and iOS decisions (`ARCHITECTURE.md`, `IOS.md`) |
| `my-money-recap/` | The illustration artwork and the scripts that produce it |
| `my-money-recap-ds/` | `@my-money-recap/ui`, the React design system |
| `apps/` | The Go backend and the SwiftUI app |
| `packages/contract/` | The Buf API contract every client is generated from |
| `infra/` | The Compose stack, the Keycloak realm, the edge, the secrets and the setup wizard |

## Getting started

```bash
npm install
```

```bash
pipx install pre-commit && pre-commit install --install-hooks
```

Then run the backend with `infra/compose/README.md` ("Run it") and the app with `apps/ios/README.md` ("Open").

## License

MIT, in `LICENSE`. The Fredoka and Nunito fonts (`apps/ios/Resources/Fonts/`) are under the SIL Open Font License, and the CLDR currency data (`apps/backend/internal/cldr/`) under the Unicode License.
