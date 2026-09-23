import { X, col, row, page, homeHeader, homeButtons, previewCard, targetCardRow, logicCommon } from './lib.mjs';

const homeInner = `${homeHeader('September 2026', 'Maya')}
  <div style="margin-top:18px">${X('Card', '', `${X('StatHero', 'eyebrow="Spent so far" value="{{ n1190 }}" format="{{ gbp }}"', '', '100%,100px')}${X('StatChipRow', '', `${X('StatChip', 'tone="earned"', 'Money in £2,400', 'auto,34px')}${X('StatChip', 'tone="saved"', 'Saved £1,210', 'auto,34px')}`, '100%,34px')}`, '100%,191px')}</div>
  <div style="margin-top:12px">${X('Card', '', `${targetCardRow('Saving target', '£1,210 of £480')}${X('ProgressBar', 'value="{{ n480 }}" max="{{ n480 }}" tone="green"', '', '100%,22px')}${X('HelperText', '', "Keep spending under £146 a day and you'll make it.", '100%,20px')}`, '100%,136px')}</div>
  <div style="margin-top:20px;display:flex;gap:12px">${X('Button', 'icon="{{ plusIcon }}"', 'Log a spend', '100%,52px')}${X('Button', 'variant="secondary" state="disabled"', 'See last month', '100%,52px')}</div>
  ${previewCard('Your first recap arrives on 1 October.')}`;
const ipad = (w, h, inner, overlay = '') => `<div style="position:relative;width:${w}px;height:${h}px;background:var(--mmr-color-ground-recap);border-radius:20px;overflow:hidden;display:flex;justify-content:center">
  ${X('Screen', `ground="recap" dc-props="{{ ${h === 1194 ? 'tallScreen' : 'shortScreen'} }}"`, inner, `600px,${h}px`)}
  ${overlay}
</div>`;
const x1 = col('X1 · iPad portrait <span style="color:var(--mmr-palette-ink-500)">· 834 × 1194</span>', ipad(834, 1194, homeInner), [
  `D40 as amended by D68: the same single-column layout, capped at 600pt on iPad instead of 390, centred on a full-bleed violet ground. Every card and field stretches to the column; the two buttons sit side by side, the one place the extra width changes composition. Type sizes, card heights and spacing are the phone's. The column runs the full height, so the recap section sits at the bottom as on the phone.`], 834);
const sheet = `<div style="position:absolute;inset:0;background:var(--mmr-color-surface-scrim);display:flex;justify-content:center;align-items:flex-end;z-index:10">
    <div style="width:600px">${X('Sheet', 'title="Log a spend"', `<div style="display:flex;flex-direction:column;gap:14px;margin-top:14px">
      ${X('AmountField', 'value="12.40" currency="GBP" label="Amount" state="focus" on-change="{{ noop }}"', '', '100%,72px')}
      <div style="display:flex;flex-wrap:wrap;gap:8px"><sc-for list="{{ catsSnacks }}" as="c" hint-placeholder-count="8">${X('Chip', 'tone="{{ c.tone }}" selected="{{ c.selected }}" tappable="{{ true }}"', `${X('Icon', 'name="{{ c.key }}"', '', '16px,16px')}{{ c.label }}`, 'auto,44px')}</sc-for></div>
      ${X('ListRow', 'label="Date" value="Today, 26 September" trailing="chevron"', '', '100%,56px')}
      ${X('TextField', 'label="Note" default-value="Coffee and lunch" max-length="{{ n80 }}"', '', '100%,72px')}
      ${X('Button', '', 'Save', '100%,52px')}
    </div>`, '390px,auto')}</div>
  </div>`;
const x2 = col('X2 · iPad landscape <span style="color:var(--mmr-palette-ink-500)">· 1194 × 834</span>', ipad(1194, 834, homeInner, sheet), [
  `D40 / §11: the same 600pt column, centred; the scrim covers the whole screen but the Sheet anchors to the bottom of the column, not the screen edge, so it takes the column's width and keeps its radius.card corners. Landscape is 10pt shorter than the phone, so the recap section tucks under the buttons.`, 'Category chips are the shipped Chip tone variant (30pt uppercase pill, 44pt hit area via tappable) as conventions.md requires. The drawn B2/B3 sheets hand-styled them as 44pt mixed-case pills; one of the two has to give; drawn here as shipped.'], 1194);


export default {
  props: { showAnnotations: { editor: 'boolean', default: true, tsType: 'boolean', section: 'Canvas' } },
  logic: `  renderVals() {${logicCommon}
    const CATS = [['snacks', 'Snacks', 'coral'], ['games', 'Games', 'violet'], ['toys', 'Toys', 'yellow'], ['rides', 'Rides', 'sky'], ['clothes', 'Clothes', 'green'], ['bills', 'Bills', 'ink'], ['gifts', 'Gifts', 'gold'], ['other', 'Other', 'rust']];
    return { ...common, catsSnacks: CATS.map(([key, label, tone]) => ({ key, label, tone, selected: key === 'snacks' })),
      tallScreen: { style: { height: 1194, width: 600 } }, shortScreen: { style: { height: 834, width: 600 } }, n80: 80, n1190: 1190, n1240: 1240 };
  }`,
  body: page('My Money Recap · iPad (X1, X2)',
    'D40 and §11: iPad uses the phone layout (one column, centred on a full-bleed ground, portrait and landscape) with the column cap raised from 390 to 600pt (D68) so the width is not wasted. A two-column iPad layout was sketched and rejected: it needs a second navigation model for little gain. These two artboards are the written rule made visible, not a second layout; nothing else is redrawn for iPad. Sheets anchor to the column. Yellow stickies are annotations, not UI.',
    [row('Row 1 · Home, portrait and landscape (D40, column widened to 600pt)', [x1, x2])]),
};
