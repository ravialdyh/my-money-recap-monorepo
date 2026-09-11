import { X, col, row, page, screen, withClose, homeHeader, homeButtons, previewCard, logicCommon } from './lib.mjs';

const heroCard = `<div style="margin-top:18px">
  ${X('Card', '', `${X('StatHero', 'eyebrow="Spent so far" value="{{ spentIdr }}" format="{{ idr }}"', '', '100%,100px')}
    ${X('StatChipRow', '', `${X('StatChip', 'tone="earned"', 'Money in Rp 15.000.000', 'auto,34px')}${X('StatChip', 'tone="saved"', 'Saved Rp 5.600.000', 'auto,34px')}`, '100%,76px')}`, '100%,233px')}
  </div>`;

const targetCard = `<div style="margin-top:12px">
  ${X('Card', '', `<div style="display:flex;flex-direction:column;gap:4px">${X('CardTitle', '', 'Saving target', 'auto,22px')}<span style="font:800 14px/1.2 var(--mmr-font-family-text);color:var(--mmr-color-text-secondary);font-variant-numeric:tabular-nums">Rp 5.600.000 of Rp 3.000.000</span></div>
    ${X('ProgressBar', 'value="{{ savedIdr }}" max="{{ targetIdr }}" tone="green"', '', '100%,22px')}
    ${X('HelperText', '', "Keep spending under Rp 520.000 a day and you'll make it.", '100%,20px')}`, '100%,157px')}
  </div>`;

const x3 = col('X3 · IDR proof · Home', screen('recap', `${homeHeader('September 2026', 'Budi', 'pandaSrc')}
  ${heroCard}
  ${targetCard}
  ${homeButtons(true)}
  ${previewCard('Your first recap arrives on 1 October.')}`), [
  'Budi, 26 September: Salary Rp 15.000.000, Payday 25, Target 20% = Rp 3.000.000, Spent Rp 9.400.000. Money In Rp 15.000.000, Saved = 15.000.000 − 9.400.000 = Rp 5.600.000. Daily line = floor((15.000.000 − 3.000.000 − 9.400.000) ÷ 5 days left) = Rp 520.000. Separators follow the phone\'s language (here id-ID: dots for thousands); the decimals follow the currency\'s CLDR data, and a whole amount drops its zero fraction (D71).',
  '<b>What IDR changes.</b> "Rp 9.400.000" does not fit the card at 80pt or 56pt, so StatHero draws it at 40pt (measured, PRD §11). The two stat chips no longer fit on one line and wrap, so the hero card grows from 191 to 233pt — which is why Home measures the card and passes PromptCard `height`. In the target card the "{Saved} of {target}" value would collide with the title, so the row wraps and the card grows 136 → 157pt. Nothing else moves.',
  'Everything here is the shipped component: StatHero measures its printed value and shrinks only if it does not fit, the Home screen passes PromptCard `height={233}` for the measured card, and every screen takes `currency` (PRD D71).',
]);

const ladder = col('X3 · StatHero fit-to-width', `<div style="display:flex;flex-direction:column;gap:12px;width:390px;padding:24px;box-sizing:border-box;background:var(--mmr-color-ground-recap);border-radius:28px">
  ${X('Card', '', `${X('StatHero', 'eyebrow="Fits at 80pt" value="{{ n0 }}" format="{{ idr }}"', '', '100%,100px')}`, '100%,140px')}
  ${X('Card', '', `${X('StatHero', 'eyebrow="Too wide at 80, fits at 56pt" value="{{ dailyIdr }}" format="{{ idr }}"', '', '100%,80px')}`, '100%,120px')}
  ${X('Card', '', `${X('StatHero', 'eyebrow="Too wide at 56, drawn at 40pt" value="{{ moneyInIdr }}" format="{{ idr }}"', '', '100%,60px')}`, '100%,100px')}
</div>`, [
  'The three sizes against the card\'s 298pt inner width, chosen by measuring the printed string: "Rp 0" fits at 80; "Rp 520.000" is wider than 298pt at 80 and fits at 56; "Rp 15.000.000" needs 40. No character counting — the same string can fit or not depending on the font and the card, so the component measures. AmountField does the same over 48 → 36 → 28 (PRD §11, D71).',
]);

const x4 = col('X4 · IDR proof · Recap cover', withClose(screen('recap', `${X('ChipRow', '', `${X('Chip', 'as-label="{{ true }}"', 'September 2026', 'auto,30px')}${X('Chip', 'as-label="{{ true }}" variant="highlight"', "Budi's recap", 'auto,30px')}`, '100%,30px')}
  <div style="margin-top:22px">${X('Headline', 'size="l"', 'Your September in money', '100%,98px')}</div>
  ${X('Illustration', 'src="{{ coverSrc }}" width="{{ n260 }}"', '', '100%,220px')}
  ${X('Card', 'under-illustration="{{ true }}"', `${X('StatHero', 'eyebrow="Money that came your way" value="{{ moneyInIdr }}" format="{{ idr }}"', '', '100%,60px')}
    ${X('StatChipRow', '', `${X('StatChip', 'tone="saved"', 'Saved Rp 5.600.000', 'auto,34px')}${X('StatChip', 'tone="spent"', 'Spent Rp 9.400.000', 'auto,34px')}`, '100%,76px')}
    ${X('HelperText', '', 'Swipe to see where it all went.', '100%,20px')}`, '100%,250px')}
  <div style="margin-top:18px">${X('PageDots', 'count="{{ n4 }}" active="{{ n0 }}" on-select="{{ noop }}"', '', 'auto,6px')}</div>`)), [
  'C1 with Budi\'s ledger, i.e. `<RecapCoverScreen currency="IDR" … />`; drawn from the same primitives here only to show the 260pt cover illustration the taller card needs.',
  '"Rp 15.000.000" needs 40pt to fit the card. The Saved and Spent chips wrap to two lines, which makes the card 42pt taller; to keep the page dots on the 844pt screen the cover illustration drops from 310 to 260pt wide. That is the one cover layout change IDR forces.',
  'D46: 44pt close control top-left (`plus` rotated 45°), the chip row shifted 34pt right to clear it. PageDots take `onSelect` and are tappable.',
]);

const x5 = col('X5 · IDR proof · Share image <span style="color:var(--mmr-palette-ink-500)">· 1080 × 1920, shown at 50%</span>', `<div style="width:540px;height:960px;overflow:hidden;flex:none">
<div style="width:1080px;height:1920px;transform:scale(.5);transform-origin:top left;box-sizing:border-box;background:var(--mmr-color-ground-personality);color:var(--mmr-color-text-on-ground);display:flex;flex-direction:column;align-items:center;padding:140px 96px 96px;font-family:var(--mmr-font-family-text)">
  <img src="{{ squirrelSrc }}" alt="" style="width:720px;max-height:620px;object-fit:contain;display:block">
  <div style="margin-top:24px;font:700 44px/1.2 var(--mmr-font-family-text);opacity:.9">This month, Budi, you were</div>
  <div style="margin-top:4px;font:700 168px/1 var(--mmr-font-family-display);letter-spacing:-.01em;text-align:center">The Squirrel</div>
  <div style="margin-top:28px;font:800 44px/1.2 var(--mmr-font-family-text);opacity:.9">Budi · September 2026</div>
  <div style="margin-top:72px;width:100%;background:var(--mmr-color-surface-card);color:var(--mmr-color-text-primary);border-radius:56px;padding:56px 72px;box-sizing:border-box;display:flex;flex-direction:column;gap:0">
    <div style="display:flex;align-items:center;justify-content:space-between;gap:32px;padding-bottom:28px;border-bottom:3px solid var(--mmr-palette-ink-100)">
      <div style="display:flex;align-items:center;gap:16px;font:800 30px/1.2 var(--mmr-font-family-text);letter-spacing:.06em;text-transform:uppercase;color:var(--mmr-color-text-secondary)"><span style="width:18px;height:18px;border-radius:999px;background:var(--mmr-color-data-saved);display:inline-block"></span>Saved</div>
      <div style="font:700 84px/1 var(--mmr-font-family-display);font-variant-numeric:tabular-nums">Rp 5.600.000</div>
    </div>
    <div style="display:flex;align-items:center;justify-content:space-between;gap:32px;padding-top:28px">
      <div style="display:flex;align-items:center;gap:16px;font:800 30px/1.2 var(--mmr-font-family-text);letter-spacing:.06em;text-transform:uppercase;color:var(--mmr-color-text-secondary)"><span style="width:18px;height:18px;border-radius:999px;background:var(--mmr-color-data-spent);display:inline-block"></span>Spent</div>
      <div style="font:700 84px/1 var(--mmr-font-family-display);font-variant-numeric:tabular-nums">Rp 9.400.000</div>
    </div>
  </div>
  <div style="margin-top:auto;font:800 34px/1.2 var(--mmr-font-family-text);letter-spacing:.12em;text-transform:uppercase;opacity:.85">My Money Recap</div>
</div>
</div>`, [
  'C6 with Budi\'s ledger: Squirrel (Saved Rp 5.600.000 ≥ Rp 3.000.000, streak assumed ≥ 7 — the IDR ledger in PRD §8 gives no streak). Archetype, month, name, Saved and Spent only; no Salary, no Money In.',
  '<b>What IDR changes.</b> The GBP poster sets Saved and Spent side by side at 120px in two 348px columns. "Rp 5.600.000" at 120px is about 720px wide, so in IDR the stats stack — one per row, label left, number right — at 84px (the poster\'s 56pt step, ×1.5). Rule: stack and step down whenever either stat is too wide for its column, measured; otherwise the two-column layout stands. The mascot is capped at 620px tall so a taller one (Hustler) cannot push the card off the poster.',
], 540);

export default {
  props: { showAnnotations: { editor: 'boolean', default: true, tsType: 'boolean', section: 'Canvas' } },
  logic: `  renderVals() {${logicCommon}
    const spentIdr = 9400000, moneyInIdr = 15000000, savedIdr = 5600000, targetIdr = 3000000, dailyIdr = 520000;
    return { ...common, spentIdr, moneyInIdr, savedIdr, targetIdr, dailyIdr, n260: 260,
    };
  }`,
  body: page('My Money Recap · IDR proofs (X3–X5)',
    'D45: every screen to date is in pounds, and "Rp 3.000.000" is three times the width of "£900". Home, the Recap cover and the share image are drawn here with the IDR ledger from PRD §8 — Budi, Salary Rp 15.000.000, Payday 25, Target 20% = Rp 3.000.000, Spent Rp 9.400.000 — before any other new screen is drawn. Where a layout has to change to fit, the change is stated on the sticky and carried into every later artboard. Yellow stickies are annotations, not UI.',
    [row('Row 1 · The three proofs', [x3, ladder, x4, x5])]),
};
