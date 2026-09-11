# 2026-09-11 canvases

Seven canvases: the IDR proofs first, then the month-so-far recap, then the four app areas, then iPad. Every screen answers a row of `../../../../SCREEN-STATES.md` Live, editable copy: <https://claude.ai/code/artifact/c2140f85-a23f-4979-a0a5-b1bdd1d2947c> (one page per file, same order).

| File | Page | Screens |
|---|---|---|
| `IDR proofs.dc.html` | 1 | X3, X4, X5 and the StatHero step-down ladder |
| `Month so far and Partial Recap.dc.html` | 2 | P1–P4 for situation 1; P1 and P4 for situations 2 and 3 |
| `Auth and Onboarding (revised).dc.html` | 3 | A0b ×2, A1 expired, A3 ×4, A5 empty, A6, A6 cap, A7 blocked, A8 ×2, flow |
| `Home and Core Features (revised).dc.html` | 4 | B1 ×8, B2 cap, B6 ×3, B7 old spend, flow |
| `Recap and Share (revised).dc.html` | 5 | C1–C4 with the close control, every C2/C3/C4 variant, C6 ×3, C7 empty, the D66 accessibility demo, flow |
| `Settings and Account (revised).dc.html` | 6 | D1 ×2, D3 ×2, D4, D6 ×2, D9 ×2, D11b, flow |
| `iPad.dc.html` | 7 | X1, X2 at the 600pt column (D68) |

Same conventions as the four canvases in the parent folder: every screen is a `<div data-screen-label>` around a 390×844 `MMR.Screen` built from `@my-money-recap/ui`; yellow stickies are annotations, not UI; the `_ds/…` references resolve only inside the Claude Design project, so these files are the record, not a local preview. The logic classes hold layout glue and sample values only — **PRD §8 is the only source of the formulas**, and every number on a sticky was checked against its worked examples.

`build/` holds the sources the canvases were generated from (`node build.mjs <Name>` writes `out/<Name>.dc.html` with the design-system bundle inlined, which is what the live copy runs). `canvas.json` is the page layout.

The decisions these screens forced are PRD D68–D71.
