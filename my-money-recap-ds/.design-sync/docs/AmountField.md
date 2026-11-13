---
category: Components
---

Currency-aware amount entry: a `TextField` with a fixed symbol and the Currency's decimals, both
from the platform formatter for `currency` and `locale`. The app takes the decimals from the server
instead, from the CLDR data `ledger` rounds with, since the platform's count differs between iOS
versions (PRD §7, D71). `value` and `onChange` carry the amount as an exact decimal string
("150000000", "12.40"), the same form it is stored in. The value is Fredoka 48px and steps to 36 and
28px when the symbol plus value is wider than the field, measured, not counted, so an 11-digit
rupiah salary still fits. The cap itself is the caller's: at 11 digits before the separator stop
accepting keystrokes and show "That's more than the app can store." as the `helper` with
`state="invalid"` (D57).
