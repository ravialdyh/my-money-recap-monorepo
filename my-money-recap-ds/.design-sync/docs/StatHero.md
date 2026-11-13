---
category: Components
---

The one big number on a screen. **It sizes itself to fit its card** (PRD §11, D71): the value is
drawn at 80px, or at 56px or 40px when the formatted string is wider than the card, measured on the
rendered text, never counted in characters, so "Rp 150.000.000" and "£2,400" each get the largest size
that actually fits. `currency` (ISO 4217) is required: there is no default, so a screen that forgets it
fails to compile instead of silently drawing pounds. `locale` is optional; the symbol, decimals and
separators come from the platform formatter, and a whole amount prints without a zero fraction
("£2,400", "£12.40"). Pass `format` instead of `currency` for values that are not money ("9 days"). The step is
measured on the final value, so a count-up never resizes mid-animation.
