// Shared helpers for the My Money Recap canvases. Output is .dc.html markup for the design canvas.
import { fileURLToPath } from 'node:url';
import path from 'node:path';
// The design-system bundle, built by `node .ds-sync/resync.mjs … --out ./ds-bundle` in my-money-recap-ds (gitignored).
export const DS = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../../../../../../my-money-recap-ds/ds-bundle');
export const X = (name, attrs = '', inner = '', size = 'auto,auto') =>
  `<x-import component-from-global-scope="MMR.${name}" ${attrs} hint-size="${size}">${inner}</x-import>`;
export const NEW = `<span style="display:inline-block;vertical-align:middle;background:var(--mmr-palette-ink-900);color:var(--mmr-palette-white);border-radius:999px;padding:2px 8px;font:800 10px/1.4 var(--mmr-font-family-text);letter-spacing:.08em;margin-right:6px">NEW</span>`;
export const REDRAW = `<span style="display:inline-block;vertical-align:middle;background:var(--mmr-palette-violet-500);color:var(--mmr-palette-white);border-radius:999px;padding:2px 8px;font:800 10px/1.4 var(--mmr-font-family-text);letter-spacing:.08em;margin-right:6px">REDRAW</span>`;
export const note = (text) => `<div style="background:var(--mmr-palette-yellow-100);border-radius:12px;padding:12px 14px;font:700 12px/1.45 var(--mmr-font-family-text)">${text}</div>`;
export const notes = (arr) => `<sc-if value="{{ showNotes }}" hint-placeholder-val="{{ true }}"><div style="display:flex;flex-direction:column;gap:8px">${arr.map(note).join('\n')}</div></sc-if>`;
export const col = (label, inner, noteArr = [], width = 390) =>
  `<!-- ${label.replace(/<[^>]+>/g, '')} -->\n<div data-screen-label="${label.replace(/<[^>]+>/g, '').replace(/"/g, '&quot;')}" style="display:flex;flex-direction:column;gap:14px;width:${width}px;flex:none">\n<div style="font:800 14px/1.2 var(--mmr-font-family-text)">${label}</div>\n${inner}\n${notes(noteArr)}\n</div>`;
export const row = (title, cols) => `<div style="display:flex;flex-direction:column;gap:24px">\n<div style="font:800 12px/1.2 var(--mmr-font-family-text);letter-spacing:.08em;text-transform:uppercase;color:var(--mmr-palette-ink-500)">${title}</div>\n<div style="display:flex;gap:80px;align-items:flex-start">\n${cols.join('\n\n')}\n</div>\n</div>`;
export const page = (title, intro, rows) => `<div style="padding:56px 80px 120px;display:flex;flex-direction:column;gap:96px;font-family:var(--mmr-font-family-text);color:var(--mmr-palette-ink-900)">\n<div style="display:flex;flex-direction:column;gap:6px;max-width:900px">\n  <div style="font:700 34px/1.1 var(--mmr-font-family-display)">${title}</div>\n  <div style="font:700 15px/1.5 var(--mmr-font-family-text);color:var(--mmr-palette-ink-500);text-wrap:pretty">${intro}</div>\n</div>\n${rows.join('\n\n')}\n</div>`;
export const screen = (ground, inner, extra = '') => X('Screen', `ground="${ground}" ${extra}`, inner, '390px,844px');
// Chrome pieces, copied from the existing canvases.
const chevronSvg = `<svg width="12" height="20" viewBox="0 0 12 20" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M10 2 2 10l8 8"></path></svg>`;
export const backRow = (chipLabel, color = 'var(--mmr-palette-white)') => `<div style="display:flex;align-items:center;gap:4px;margin-left:-14px">\n    <button aria-label="Back" style="width:44px;height:44px;border:0;background:transparent;color:${color};display:flex;align-items:center;justify-content:center;cursor:pointer;padding:0">${chevronSvg}</button>\n    ${X('Chip', 'as-label="{{ true }}"', chipLabel, 'auto,30px')}\n  </div>`;
export const sheetBack = `<button aria-label="Back" style="width:44px;height:44px;margin-left:-14px;border:0;background:transparent;color:var(--mmr-palette-ink-900);display:flex;align-items:center;justify-content:center;cursor:pointer;padding:0">${chevronSvg}</button>`;
// D46 close control: 44pt, `plus` rotated 45°, top-left, laid over the recap screen; the wrapper shifts the first chip row right by 34pt.
export const closeBtn = (dark = false) => `<button aria-label="Close" style="position:absolute;left:10px;top:49px;width:44px;height:44px;border:0;background:transparent;color:${dark ? 'var(--mmr-palette-ink-900)' : 'var(--mmr-palette-white)'};display:flex;align-items:center;justify-content:center;cursor:pointer;padding:0;z-index:3"><span style="display:inline-flex;transform:rotate(45deg)">${X('Icon', 'name="plus"', '', '20px,20px')}</span></button>`;
export const withClose = (screenMarkup, dark = false) => `<div class="mmr-with-close" style="position:relative;width:390px;height:844px">${closeBtn(dark)}${screenMarkup}</div>`;
// Home header (B1): month chip + settings chip, greeting + avatar.
export const homeHeader = (month, name, avatarKey = 'avatarSrc', opts = {}) => `${X('ChipRow', '', `${X('Chip', `tappable="{{ true }}"${opts.monthState ? ` state="${opts.monthState}"` : ''}`, month, 'auto,30px')}<div style="flex-grow:1"></div>${X('Chip', 'variant="inverse" tappable="{{ true }}"', `${X('Icon', 'name="settings" size="{{ n14 }}"', '', '14px,14px')}<span style="margin-left:6px">Settings</span>`, 'auto,30px')}`, '100%,30px')}
  ${opts.after || ''}
  <div style="display:flex;justify-content:space-between;align-items:center;margin-top:18px">
    ${X('Headline', 'size="m"', `Hi, ${name}`, 'auto,42px')}
    ${X('Avatar', `src="{{ ${avatarKey} }}" label="Avatar" size="{{ n56 }}"`, '', '56px,56px')}
  </div>`;
export const targetCardRow = (title, valueText) => `<div style="display:flex;justify-content:space-between;align-items:baseline">${X('CardTitle', '', title, 'auto,22px')}<span style="font:800 14px/1.2 var(--mmr-font-family-text);color:var(--mmr-color-text-secondary);font-variant-numeric:tabular-nums">${valueText}</span></div>`;
export const homeButtons = (seeLastDisabled = true, logDisabled = false) => `<div style="margin-top:20px;display:flex;flex-direction:column;gap:12px">
    ${X('Button', `icon="{{ plusIcon }}"${logDisabled ? ' state="disabled"' : ''}`, 'Log a spend', '100%,52px')}
    ${X('Button', `variant="secondary"${seeLastDisabled ? ' state="disabled"' : ''}`, 'See last month', '100%,52px')}
  </div>`;
// D47 preview card: dashed, tappable, chevron well, opens P1.
export const previewCard = (text, label = 'Past recaps') => `<div style="margin-top:auto;display:flex;flex-direction:column;gap:8px">
    <div style="font:800 12px/1.2 var(--mmr-font-family-text);letter-spacing:.08em;text-transform:uppercase;opacity:.85">${label}</div>
    <button style="display:flex;align-items:center;justify-content:space-between;gap:12px;width:100%;border:2px dashed rgba(255,255,255,.4);border-radius:28px;background:transparent;padding:10px 12px 10px 22px;font:700 14px/1.4 var(--mmr-font-family-text);color:var(--mmr-color-text-on-ground-muted);text-align:left;cursor:pointer;min-height:56px"><span>${text}</span><span style="width:36px;height:36px;border-radius:999px;background:rgba(255,255,255,.14);color:var(--mmr-palette-white);display:inline-flex;align-items:center;justify-content:center;flex:none">${X('Icon', 'name="chevron"', '', '20px,20px')}</span></button>
  </div>`;
export const flowNode = `<div style="position:absolute;left:{{ f.x }};top:{{ f.y }};width:96px;display:flex;flex-direction:column;gap:8px;align-items:center">
      <div style="width:96px;height:208px;border-radius:12px;background:{{ f.bg }};position:relative;overflow:hidden">
        <div style="position:absolute;left:{{ f.inset }};right:{{ f.inset }};top:{{ f.sheetTop }};bottom:{{ f.inset }};border-radius:8px;background:var(--mmr-palette-white);opacity:{{ f.sheetOn }}"></div>
        <div style="position:absolute;left:10px;top:16px;width:52px;height:8px;border-radius:999px;background:var(--mmr-palette-white);opacity:{{ f.pills }}"></div>
        <div style="position:absolute;left:10px;top:30px;width:36px;height:8px;border-radius:999px;background:var(--mmr-palette-white);opacity:{{ f.pillsDim }}"></div>
        <div style="position:absolute;left:{{ f.cardX }};right:{{ f.cardX }};top:{{ f.cardTop }};bottom:{{ f.cardBottom }};border-radius:8px;background:{{ f.cardBg }}"></div>
        <div style="position:absolute;left:8px;right:8px;bottom:8px;height:14px;border-radius:999px;background:{{ f.btnBg }};opacity:{{ f.btn }}"></div>
        <div style="position:absolute;inset:0;background:var(--mmr-color-surface-scrim);opacity:{{ f.dim }}"></div>
        <div style="position:absolute;left:0;right:0;bottom:0;height:{{ f.sheetH }};border-radius:10px 10px 0 0;background:var(--mmr-palette-white);border:1.5px solid var(--mmr-palette-ink-100);border-bottom:0;box-sizing:border-box"></div>
        <div style="position:absolute;left:14px;right:14px;top:80px;height:48px;border-radius:8px;background:var(--mmr-palette-white);opacity:{{ f.alert }}"></div>
        <div style="position:absolute;left:8px;right:8px;top:8px;height:22px;border-radius:6px;background:{{ f.bannerBg }};opacity:{{ f.banner }}"></div>
      </div>
      <div style="font:800 11px/1.3 var(--mmr-font-family-text);color:var(--mmr-palette-ink-900);text-align:center;text-wrap:balance">{{ f.label }}</div>
    </div>`;
export const flowLabel = `<div style="position:absolute;left:{{ l.x }};top:{{ l.y }};transform:translate(-50%,-50%);background:var(--mmr-palette-white);padding:3px 8px;border-radius:999px;font:800 11px/1.2 var(--mmr-font-family-text);color:{{ l.color }};white-space:nowrap">{{ l.text }}</div>`;
export const flow = (label, w, h, svgPaths, markerId, caption, nodeCount, labelCount, extra = '', captionPos = 'left:60px;top:28px;width:520px') => `<div data-screen-label="${label}" style="display:flex;flex-direction:column;gap:14px;width:${w}px">
<div style="font:800 14px/1.2 var(--mmr-font-family-text)">${label}</div>
<div style="position:relative;width:${w}px;height:${h}px;background:var(--mmr-palette-white);border-radius:28px;overflow:hidden">
  <svg width="${w}" height="${h}" viewBox="0 0 ${w} ${h}" fill="none" stroke="var(--mmr-palette-ink-500)" stroke-width="2" style="position:absolute;inset:0" aria-hidden="true">
    <defs><marker id="${markerId}" markerWidth="8" markerHeight="8" refX="6" refY="4" orient="auto"><path d="M0 0 8 4 0 8Z" fill="var(--mmr-palette-ink-500)" stroke="none"></path></marker></defs>
    ${svgPaths.map(d => `<path d="${d}" marker-end="url(#${markerId})"></path>`).join('\n    ')}
  </svg>
  ${extra}
  <sc-for list="{{ flowNodes }}" as="f" hint-placeholder-count="${nodeCount}">${flowNode}</sc-for>
  <sc-for list="{{ flowLabels }}" as="l" hint-placeholder-count="${labelCount}">${flowLabel}</sc-for>
  <div style="position:absolute;${captionPos};font:700 13px/1.4 var(--mmr-font-family-text);color:var(--mmr-palette-ink-500);text-wrap:pretty">${caption}</div>
</div>
</div>`;
// Logic-side helpers, emitted as source text so every canvas carries the same definitions.
export const logicCommon = `
    const K = window.MMR, R = window.React, I = K ? K.ILLUSTRATIONS : {};
    const showNotes = this.props.showAnnotations ?? true;
    const gbp = (n) => (n < 0 ? '−' : '') + '£' + Math.abs(Math.round(n)).toLocaleString('en-GB');
    const idr = (n) => 'Rp ' + Math.round(n).toLocaleString('id-ID');
    const icon = (name, props = {}) => K ? R.createElement(K.Icon, { name, ...props }) : null;
    const G = { violet: 'var(--mmr-color-ground-recap)', yellow: 'var(--mmr-color-ground-earning)', coral: 'var(--mmr-color-ground-goal)', green: 'var(--mmr-color-ground-personality)' };
    // Flow thumbnails: N(x, y, label, options)
    const N = (x, y, label, o = {}) => ({ x: x + 'px', y: y + 'px', label, bg: o.bg || G.violet,
      inset: o.sheet ? '6px' : '0px', sheetTop: o.sheet ? '14px' : '0px', sheetOn: o.sheet ? 1 : 0, pills: o.sheet ? 0 : 1, pillsDim: o.sheet ? 0 : 0.6,
      cardX: o.sheet ? '14px' : '8px', cardTop: o.cardTop ?? '56px', cardBottom: o.cardBottom ?? '96px', cardBg: o.sheet ? 'var(--mmr-color-surface-well-violet)' : 'var(--mmr-palette-white)',
      btnBg: o.coral ? 'var(--mmr-palette-coral-500)' : 'var(--mmr-palette-ink-900)', btn: o.btn ?? 0, dim: o.dim ?? 0, sheetH: o.sheetH ?? '0px', alert: o.alert ?? 0,
      banner: o.banner ? 1 : 0, bannerBg: o.banner === 'grey' ? 'var(--mmr-palette-ink-500)' : 'var(--mmr-palette-coral-500)' });
    const L = (x, y, text, color) => ({ x: x + 'px', y: y + 'px', text, color: color || 'var(--mmr-palette-ink-900)' });
    const common = {
      showNotes, gbp, idr, plusIcon: icon('plus'), shareIcon: icon('share'), piggy: icon('piggy'),
      avatarSrc: I['avatar-otter'], pandaSrc: I['avatar-panda'], coverSrc: I.cover, jarSrc: I.jar, vendingSrc: I.vending, envelopeSrc: I.envelope,
      squirrelSrc: I.squirrel, hustlerSrc: I.hustler, treatSrc: I.treat, slowburnerSrc: I.slowburner,
      n0: 0, n1: 1, n2: 2, n3: 3, n4: 4, n14: 14, n56: 56, n30: 30, n31: 31, n480: 480,
      noop: () => {},
      onGround: { style: { color: 'var(--mmr-color-text-on-ground-muted)' } },
      tightCard: { style: { padding: '4px 22px' } },
      fullSheet: { style: { height: 'auto', flexGrow: 1 } },
      tintCard: { style: { gap: 12 } },
      gap12: { style: { gap: 12 } }, gap14: { style: { gap: 14 } },
      growCard: { style: { flexGrow: 1 } },
      jarStyle: { style: { marginBottom: -10 } },
    };`;
export const extraCss = `
body{margin:0;background:#EFECF7;font-family:Nunito,'Trebuchet MS',sans-serif;color:#1E1B3A}
a{color:#5B4FE8}a:hover{color:#1E1B3A}
input::placeholder{color:#6B6890;font-weight:700}
/* D46 close control: the first chip row of a recap screen moves right 34pt to clear the 44pt control. */
.mmr-with-close .mmr-screen > .mmr-chip-row:first-child{margin-left:34px}
/* B1 first load skeleton (D49): ink.100 blocks in place of numbers. */
.mmr-skeleton{background:var(--mmr-palette-ink-100);border-radius:999px;display:inline-block}
`;
export const a11yCss = `
/* D66 demonstration: the "Large" accessibility size, about 1.35× the default ramp. Layout becomes one vertical scroll. */
.mmr-a11y .mmr-screen{height:auto;min-height:0;padding-top:28px;padding-bottom:28px}
.mmr-a11y .mmr-headline[data-size="l"]{font-size:64px}
.mmr-a11y .mmr-headline[data-size="m"]{font-size:54px}
.mmr-a11y .mmr-headline[data-size="compact"]{font-size:48px}
.mmr-a11y .mmr-headline[data-size="display"]{font-size:90px}
.mmr-a11y .mmr-subcopy{font-size:22px}
.mmr-a11y .mmr-helper,.mmr-a11y .mmr-eyebrow,.mmr-a11y .mmr-field-label{font-size:19px}
.mmr-a11y .mmr-card-title,.mmr-a11y .mmr-sheet-title{font-size:27px}
.mmr-a11y .mmr-chip{font-size:16px;padding:10px 18px}
.mmr-a11y .mmr-stat-chip{font-size:18px;padding:12px 18px}
.mmr-a11y .mmr-bar-labels,.mmr-a11y .mmr-compare-label,.mmr-a11y .mmr-compare-value,.mmr-a11y .mmr-row-label,.mmr-a11y .mmr-list-row-value{font-size:18px}
.mmr-a11y .mmr-compare-label{width:74px}.mmr-a11y .mmr-compare-value{width:80px}
.mmr-a11y .mmr-button{font-size:22px;height:64px}
.mmr-a11y .mmr-badge{width:84px;height:84px}.mmr-a11y .mmr-badge-value{font-size:40px}.mmr-a11y .mmr-badge-unit{font-size:12px}
.mmr-a11y .mmr-archetype{font-size:14px;padding:8px 14px}
.mmr-a11y .mmr-illo{flex-grow:0}
`;
