---
category: Components
---

The one big number on a screen. **It sizes itself from the formatted string** (PRD §11, D45): 80px up
to 7 glyphs, 56px from 8 to 10, 40px from 11 — counted on the symbol, separators and spaces too, because
that is what has to fit. "£2,400" is 6 glyphs; "Rp 15.000.000" is 13. Pass `format` for the Member's
Currency; the default is pounds. The step is measured on the final value, so a count-up never resizes
mid-animation.
