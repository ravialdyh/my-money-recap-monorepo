---
category: Components
---

Currency-aware amount entry with a fixed symbol and minor-unit formatting (IDR 0, GBP 2, BHD 3). The
value is Fredoka 48px and steps to 36 and 28 as the symbol plus the integer part passes 7 and 10 glyphs
(PRD §11) — the fraction is fixed per Currency and does not count — so "£ 2,400" stays at 48 and a
12-digit rupiah salary still fits the card. The cap itself is the caller's: at 12 digits before the
separator stop accepting keystrokes and show "That's more than the app can store." as the `helper` with
`state="invalid"` (D57).
