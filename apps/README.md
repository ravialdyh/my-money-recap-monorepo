# apps

Runnable products. Nothing here yet.

Planned, per `docs/ARCHITECTURE.md`: the iOS app (`apps/ios`), the Go BFF (`apps/bff`)
and the four Go services (`apps/members`, `apps/ledger`, `apps/notify`, `apps/currency`).
Each gets its own folder. The iOS app consumes the design tokens and illustrations from
`my-money-recap-ds`; the Go apps consume the generated contract from `packages/contract`.
