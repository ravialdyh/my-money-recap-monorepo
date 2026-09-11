import { X, col, row, page, screen, backRow, homeHeader, homeButtons, previewCard, targetCardRow, flow, logicCommon, NEW, REDRAW } from './lib.mjs';

const hero = (eyebrow, valueKey, chips, h = 191) => `<div style="margin-top:18px">${X('Card', '', `${X('StatHero', `eyebrow="${eyebrow}" value="{{ ${valueKey} }}" format="{{ gbp }}"`, '', '100%,100px')}
    ${X('StatChipRow', '', chips.map(([tone, text]) => X('StatChip', `tone="${tone}"`, text, 'auto,34px')).join(''), '100%,34px')}`, `100%,${h}px`)}</div>`;
const target = (valueText, valueKey, helper, top = 12) => `<div style="margin-top:${top}px">${X('Card', '', `${targetCardRow('Saving target', valueText)}
    ${X('ProgressBar', `value="{{ ${valueKey} }}" max="{{ n480 }}" tone="green"`, '', '100%,22px')}
    ${X('HelperText', '', helper, '100%,20px')}`, '100%,136px')}</div>`;
const plainRecaps = (text = 'Your first recap arrives on 1 October.') => `<div style="margin-top:auto;display:flex;flex-direction:column;gap:8px">
    <div style="font:800 12px/1.2 var(--mmr-font-family-text);letter-spacing:.08em;text-transform:uppercase;opacity:.85">Past recaps</div>
    ${X('HelperText', 'dc-props="{{ onGround }}"', text, '100%,20px')}
  </div>`;
const pastList = (rows) => `<div style="margin-top:auto;display:flex;flex-direction:column;gap:8px">
    <div style="font:800 12px/1.2 var(--mmr-font-family-text);letter-spacing:.08em;text-transform:uppercase;opacity:.85">Past recaps</div>
    ${X('Card', 'dc-props="{{ tightCard }}"', X('DataRowList', '', rows.map(([l, v]) => X('ListRow', `label="${l}" value="${v}" trailing="chevron"`, '', '100%,56px')).join(''), '100%,112px'), '100%,120px')}
  </div>`;
const H = (month, name = 'Maya', opts) => homeHeader(month, name, 'avatarSrc', opts);

// Row 1 · Home states
const b1Before = col('B1 · Home / before payday', screen('recap', `${H('September 2026')}
  ${hero('Spent so far · salary lands in 5 days', 'n900', [['earned', 'Money in £0']])}
  ${target('£0 of £480', 'n0', 'Nothing to keep yet. Your salary lands on the 25th.')}
  ${homeButtons(true)}
  ${plainRecaps()}`), [
  `${REDRAW}D41: the "Saved −£900" chip is deleted. Before Money In > 0 the row carries only "Money in £0". 20 September, Spent £900, Money In £0 (payday the 25th). Everything else — the hero, the countdown eyebrow, the bar clamped at zero, the helper — stays exactly as drawn.`,
  'A negative Saved on a green chip stated the opposite of the Member\'s position; Spent and Money In together already say everything true. VoiceOver: the hidden chip is absent from the tree, not present-and-empty (§11).']);
const b1NoSalary = col('B1 · Home / no salary set', screen('recap', `${H('September 2026')}
  ${hero('Spent so far', 'n1190', [['earned', 'Money in £0']])}
  <div style="margin-top:12px">${X('PromptCard', 'size="hero" title="Set your monthly salary" helper="We\'ll show what you kept once we know what came in."', '', '100%,191px')}</div>
  <div style="margin-top:12px">${X('PromptCard', 'size="target" title="Set a saving target" helper="Pick a share of your salary to keep each month."', '', '100%,136px')}</div>
  ${homeButtons(true)}
  ${plainRecaps()}`), [
  `${REDRAW}D54: the hero keeps the StatHero — Spent £1,190, the spend the Member just logged — and the salary PromptCard (size="hero") sits below it. The drawn version replaced the whole hero with the prompt and hid the spend. With no Salary there is no Payday, so the eyebrow has no countdown and Money In is £0.`,
  'This is the tallest Home: hero 191 + prompt 191 + prompt 136 puts the buttons at 740–856pt, so Home scrolls here; "Log a spend" stays above the fold and the recap section sits below it. If scrolling is not wanted, the salary prompt at size="auto" (about 94pt) brings the page back under 844.']);
const b1NoTarget = col('B1 · Home / no target set, salary set', screen('recap', `${H('September 2026')}
  ${hero('Spent so far', 'n1190', [['earned', 'Money in £2,400'], ['saved', 'Saved £1,210']])}
  <div style="margin-top:12px">${X('PromptCard', 'size="target" title="Set a saving target" helper="Pick a share of your salary to keep each month."', '', '100%,136px')}</div>
  ${homeButtons(true)}
  ${plainRecaps()}`), [
  `${NEW}G-01. 26 September: the hero card is entirely normal — Spent £1,190, Money in £2,400, Saved £1,210. Only the target card becomes the 136pt PromptCard "Set a saving target", so nothing below it moves when the target is set. No daily-allowance sentence (it derives from the target amount) and the Target Warning cannot fire (§8.8).`]);
const b1WarnBefore = col('B1 · Home / target warning, before payday', screen('recap', `${H('September 2026')}
  ${hero('Spent so far · salary lands in 5 days', 'n2000', [['earned', 'Money in £0']])}
  <div style="margin-top:12px">${X('Banner', 'tone="warning" on-dismiss="{{ noop }}"', "You've spent £2,000 of the £1,920 you meant to spend this month.", '100%,72px')}</div>
  ${target('£0 of £480', 'n0', 'Nothing to keep yet. Your salary lands on the 25th.')}
  ${homeButtons(true)}`), [
  `${NEW}D53: the warning fires before Payday too, with different copy. 20 September, Spent £2,000 > the line of £2,400 − £480 = £1,920. Same coral Banner, dismiss × for the session only, target card moves up on dismiss. Money In is still £0, so no Saved chip and the bar stays clamped at zero.`]);
const b1Partial = col('B1 · Home / partial month in the list', screen('recap', `${H('October 2026')}
  ${hero('Spent so far · salary lands in 24 days', 'n0', [['earned', 'Money in £0']])}
  ${target('£0 of £480', 'n0', 'Nothing spent yet. Streak day 1.')}
  ${homeButtons(false)}
  ${pastList([['September 2026', 'The Squirrel'], ['August 2026', 'Partial']])}`), [
  `${NEW}§8.11: a partial month shows "Partial" as its value instead of an Archetype name and opens P1–P4 (situation 2 or 3). Here August was the sign-up month with the 25th already past. Newest first; the current month is never listed.`,
  'Drawn on 1 October by the PRD rules: Money In is £0 until the 25th, so no Saved chip and "salary lands in 24 days"; the window starts on the 1st, so "Streak day 1". (The drawn "with past recaps" sibling shows Money in £2,400 on 1 October, which §8.2 contradicts — left as is, since that row is DRAWN.)']);

// Row 2 · offline, first load, preview card, log-a-spend cap
const b1Offline = col('B1 · Home / offline', screen('recap', `${H('September 2026', 'Maya', { after: `<div style="margin-top:8px;font:700 12px/1.3 var(--mmr-font-family-text);color:var(--mmr-color-text-on-ground-muted)">Last updated Saturday 14:20.</div>
  <div style="margin-top:12px">${X('Banner', 'tone="offline"', 'No connection. You can look, but not log.', '100%,60px')}</div>` })}
  ${hero('Spent so far', 'n1190', [['earned', 'Money in £2,400'], ['saved', 'Saved £1,210']])}
  ${target('£1,210 of £480', 'n480', "Keep spending under £146 a day and you'll make it.")}
  ${homeButtons(true, true)}`), [
  `${NEW}D58 / D61 / D63: read offline, write online. Every number renders from the cache; the grey Banner sits at the top and "Last updated {day} {time}" is the month chip's second line, because an iPad's cache can lag an iPhone's writes. "Log a spend" is disabled at 40%; the sheet is not reachable.`,
  'Banner `tone="offline"` — the grey third tone beside warning and info.']);
const skel = (w, h, r = 999) => `<span class="mmr-skeleton" style="width:${w}px;height:${h}px;border-radius:${r}px"></span>`;
const b1FirstLoad = col('B1 · Home / first load, no cache', screen('recap', `${H('September 2026')}
  <div style="margin-top:18px">${X('Card', '', `<div style="display:flex;flex-direction:column;gap:10px">${skel(96, 14)}${skel(190, 64, 16)}</div>
    <div style="display:flex;gap:8px">${skel(128, 34)}${skel(112, 34)}</div>`, '100%,191px')}</div>
  <div style="margin-top:12px">${X('Card', '', `<div style="display:flex;justify-content:space-between;align-items:center">${X('CardTitle', '', 'Saving target', 'auto,22px')}${skel(84, 14)}</div>
    ${X('ProgressBar', 'value="{{ n0 }}" max="{{ n480 }}" tone="green"', '', '100%,22px')}
    <div style="display:flex">${skel(230, 14)}</div>`, '100%,136px')}</div>
  ${homeButtons(true)}`), [
  `${NEW}D49: no cache exists after a first sign-in. Both cards at their filled heights — 191 and 136pt — with ink.100 blocks where the numbers go, so nothing jumps when they arrive. No spinner. The greeting is known from onboarding and stays.`]);
const b1Preview = col('B1 · Home / preview card', screen('recap', `${H('September 2026')}
  ${hero('Spent so far', 'n1190', [['earned', 'Money in £2,400'], ['saved', 'Saved £1,210']])}
  ${target('£1,210 of £480', 'n480', "Keep spending under £146 a day and you'll make it.")}
  ${homeButtons(true)}
  ${previewCard('Your first recap arrives on 1 October.')}`), [
  `${REDRAW}D47: the dashed "Your first recap arrives on 1 October." card is now tappable — a chevron well on the right, 56pt minimum — and opens the month-so-far preview (P1). It replaces the plain helper line under "Past recaps" wherever no Recap exists yet.`]);
const b2TooLong = col('B2 · Log a spend / amount too long', screen('recap', `${H('September 2026')}
  <div style="position:absolute;inset:0">
  ${X('Sheet', 'title="Log a spend" scrim="{{ true }}"', `<div style="display:flex;flex-direction:column;gap:14px;margin-top:14px">
      ${X('AmountField', 'value-minor="{{ capMinor }}" currency="GBP" label="Amount" state="invalid" helper="That\'s more than the app can store." on-change="{{ noop }}"', '', '100%,96px')}
      <div style="display:flex;flex-wrap:wrap;gap:8px"><sc-for list="{{ catsSnacks }}" as="c" hint-placeholder-count="8">${X('Chip', 'tone="{{ c.tone }}" selected="{{ c.selected }}" tappable="{{ true }}"', `${X('Icon', 'name="{{ c.key }}"', '', '16px,16px')}{{ c.label }}`, 'auto,44px')}</sc-for></div>
      ${X('ListRow', 'label="Date" value="Today, 26 September" trailing="chevron"', '', '100%,56px')}
      ${X('TextField', 'label="Note" placeholder="What was it? (optional)"', '', '100%,72px')}
      ${X('Button', '', 'Save', '100%,52px')}
    </div>`, '100%,100%')}
  </div>`), [
  `${NEW}D57, same rule as A6: at 12 digits before the separator the field stops accepting keystrokes and shows "That's more than the app can store." — the line is coral but the amount is still a valid one, so Save stays enabled once a Category is picked.`,
  'AmountField counts symbol + integer part: "£ 999,999,999,999" is 17 glyphs, so it sits at the 28pt step (48 → 36 → 28, §11, D71).', 'Category chips are the shipped Chip tone variant (30pt uppercase pill, 44pt hit area via tappable) as conventions.md requires. The drawn B2/B3 sheets hand-styled them as 44pt mixed-case pills — one of the two has to give; drawn here as shipped.']);

// Row 3 · Spend list, empty list, old spend
const switcher = (open = false) => `<div style="margin-top:12px;display:flex;align-items:center;gap:8px;position:relative">
    ${X('Chip', `tappable="{{ true }}"${open ? ' state="pressed"' : ''}`, `September 2026<span style="display:inline-flex;transform:rotate(90deg)">${X('Icon', 'name="chevron" size="{{ n14 }}"', '', '14px,14px')}</span>`, 'auto,30px')}
  </div>`;
const listCard = `<div style="margin-top:20px">${X('Card', '', X('DataRowList', '', `<sc-for list="{{ catRows }}" as="r" hint-placeholder-count="6">${X('DataRow', 'icon="{{ r.icon }}" icon-tone="{{ r.tone }}" label="{{ r.label }}" amount="{{ r.amount }}"', '', '100%,62px')}</sc-for>`, '100%,380px'), '100%,420px')}</div>`;
const b6 = col('B6 · Spend list', screen('recap', `${backRow('Spends')}
  <div style="margin-top:16px">${X('Headline', 'size="m"', 'Spends', '100%,42px')}</div>
  ${switcher()}
  <div style="margin-top:10px">${X('SubCopy', '', '£1,190 so far, 11 spends', '100%,22px')}</div>
  ${listCard}`), [
  `${REDRAW}D59: a month switcher under the title, the current month by default. The title stops being the hardcoded "September spends". Tapping the switcher lists every month with data, newest first (next artboard); it is the only place the Member changes which month they are looking at, and the Home month chip opens the list here.`]);
const b6Open = col('B6 · Spend list / month switcher open', screen('recap', `${backRow('Spends')}
  <div style="margin-top:16px">${X('Headline', 'size="m"', 'Spends', '100%,42px')}</div>
  ${switcher(true)}
  <div style="margin-top:10px">${X('SubCopy', '', '£1,190 so far, 11 spends', '100%,22px')}</div>
  <div style="position:relative">
    <div style="opacity:.4">${listCard}</div>
    <div style="position:absolute;left:0;right:0;top:8px">${X('Card', 'dc-props="{{ tightCard }}"', X('DataRowList', '', `${X('ListRow', 'label="September 2026" trailing="check" selected="{{ true }}"', '', '100%,56px')}${X('ListRow', 'label="August 2026" trailing="check"', '', '100%,56px')}${X('ListRow', 'label="July 2026" trailing="check"', '', '100%,56px')}`, '100%,168px'), '100%,176px')}</div>
  </div>`), [
  `${REDRAW}The open switcher: every month with data, newest first, as ListRow trailing="check" — the same row as the currency list. Months with no Spend and no Money In are not listed. Picking a past month reloads the list; editing there recomputes that month's Recap live (D22).`]);
const b6Empty = col('B6 · Spend list / empty', screen('recap', `${backRow('Spends')}
  <div style="margin-top:16px">${X('Headline', 'size="m"', 'Spends', '100%,42px')}</div>
  ${switcher()}
  <div style="margin-top:10px">${X('SubCopy', '', '£0 so far, no spends', '100%,22px')}</div>
  <div style="flex-grow:1;display:flex;flex-direction:column;align-items:center;justify-content:center;gap:20px;text-align:center">
    <div style="font:700 22px/1.25 var(--mmr-font-family-display);text-wrap:balance;max-width:260px">Nothing spent yet. Streak day 23.</div>
    <button aria-label="Log a spend" style="width:56px;height:56px;border:0;border-radius:999px;background:var(--mmr-palette-ink-900);color:var(--mmr-palette-white);display:flex;align-items:center;justify-content:center;cursor:pointer">${X('Icon', 'name="plus"', '', '20px,20px')}</button>
  </div>`), [
  `${REDRAW}D43: the streak respects the sign-up window. 26 September for a Member who signed up on the 4th: the window is 4–26 September, 23 elapsed days, so "Streak day 23" — not the day of the month. The drawn "Streak day 4" on a board set to the 21st–30th was wrong under any rule. The switcher is present here too.`]);
const b7Old = col('B7 · Spend detail / older than 60 days', screen('recap', `${backRow('Spends')}
  <div style="margin-top:16px">${X('Headline', 'size="m"', 'Spends', '100%,42px')}</div>
  <div style="position:absolute;inset:0">
  ${X('Sheet', 'title="Edit spend" scrim="{{ true }}"', `<div style="display:flex;flex-direction:column;gap:14px;margin-top:14px">
      ${X('AmountField', 'value-minor="{{ n18000 }}" currency="GBP" label="Amount" on-change="{{ noop }}"', '', '100%,72px')}
      <div style="display:flex;flex-wrap:wrap;gap:8px"><sc-for list="{{ catsClothes }}" as="c" hint-placeholder-count="8">${X('Chip', 'tone="{{ c.tone }}" selected="{{ c.selected }}" tappable="{{ true }}"', `${X('Icon', 'name="{{ c.key }}"', '', '16px,16px')}{{ c.label }}`, 'auto,44px')}</sc-for></div>
      <div style="display:flex;flex-direction:column;gap:4px">${X('ListRow', 'label="Date" value="10 July 2026" trailing="none" state="disabled"', '', '100%,56px')}${X('HelperText', '', "Dates older than 60 days can't be changed.", '100%,20px')}</div>
      ${X('TextField', 'label="Note" default-value="Summer jacket" max-length="{{ n80 }}"', '', '100%,72px')}
      ${X('Button', '', 'Save changes', '100%,52px')}
      ${X('ListRow', 'label="Delete this spend" tone="destructive" trailing="none"', '', '100%,56px')}
    </div>`, '100%,100%')}
  </div>`), [
  `${NEW}§6.10: opened on 26 September, the picker only offers the last 60 days including today (29 July onward), so a spend dated 10 July keeps its date. The date row becomes read-only text — no chevron, 40% — with the helper "Dates older than 60 days can't be changed." Amount, Category and Note stay editable; deleting still works.`]);

const flowB = flow('Flow · Home and the daily loop', 1440, 980, [
  'M156 404H298', 'M348 300C348 144 440 144 538 144', 'M636 144H778', 'M876 144H1018', 'M876 230C925 230 925 380 878 380',
  'M778 200C700 200 560 360 398 360', 'M778 440H398', 'M348 508C348 664 440 664 538 664', 'M636 664H778', 'M876 664H1018', 'M1116 664H1258',
  'M1308 804C1308 872 828 872 828 806', 'M300 330C250 330 250 160 232 160', 'M300 480C250 480 250 860 232 860',
  'M348 250V210', 'M300 440C230 440 230 664 158 664', 'M588 520C560 560 588 600 588 556',
], 'arwB', 'Home numbers animate to their new values with motion.reveal.number after the sheet closes; Reduce Motion replaces every reveal with a 150 ms fade. New since the first canvas: the offline branch (D58 — Home reads from cache, "Log a spend" is disabled and the sheet is unreachable), the month switcher on the Spend list (D59) and the preview entry from the Home card (D47). Thumbnails are schematic; see rows 1–3 for the screens.', 12, 17,
  `<div style="position:absolute;left:40px;top:140px;width:190px;height:40px;box-sizing:border-box;border:2px dashed var(--mmr-palette-ink-100);border-radius:999px;display:flex;align-items:center;justify-content:center;font:800 11px/1.2 var(--mmr-font-family-text);color:var(--mmr-palette-ink-500)">Settings (part 4)</div>
  <div style="position:absolute;left:40px;top:840px;width:190px;height:40px;box-sizing:border-box;border:2px dashed var(--mmr-palette-ink-100);border-radius:999px;display:flex;align-items:center;justify-content:center;font:800 11px/1.2 var(--mmr-font-family-text);color:var(--mmr-palette-ink-500)">Recap (part 3)</div>
  <div style="position:absolute;left:300px;top:10px;width:96px;height:200px;box-sizing:border-box;border:2px dashed var(--mmr-palette-ink-100);border-radius:12px;display:flex;align-items:center;justify-content:center;text-align:center;padding:8px;font:800 11px/1.3 var(--mmr-font-family-text);color:var(--mmr-palette-ink-500)">P1–P4 · Month so far (new group)</div>`, 'left:1150px;top:28px;width:270px');

export default {
  props: { showAnnotations: { editor: 'boolean', default: true, tsType: 'boolean', section: 'Canvas' } },
  logic: `  renderVals() {${logicCommon}
    const CATS = [['snacks', 'Snacks', 'coral'], ['games', 'Games', 'violet'], ['toys', 'Toys', 'yellow'], ['rides', 'Rides', 'sky'], ['clothes', 'Clothes', 'green'], ['bills', 'Bills', 'ink'], ['gifts', 'Gifts', 'gold'], ['other', 'Other', 'rust']];
    const mkCats = (sel) => CATS.map(([key, label, tone]) => ({ key, label, tone, selected: sel === key }));
    const sub = (text, small) => R.createElement('span', { style: { display: 'flex', flexDirection: 'column', gap: 2 } }, text, R.createElement('span', { style: { fontSize: 12, fontWeight: 700, color: 'var(--mmr-color-text-secondary)' } }, small));
    const catRows = [['bills', 'Bills', 3, 620], ['snacks', 'Snacks', 4, 210], ['clothes', 'Clothes', 1, 180], ['gifts', 'Gifts', 1, 120], ['rides', 'Rides', 1, 40], ['other', 'Other', 1, 20]].map(([key, label, n, total]) => ({
      icon: icon(key), tone: CATS.find(c => c[0] === key)[2], label: sub(label, n + ' spend' + (n === 1 ? '' : 's')), amount: gbp(total) }));
    const flowNodes = [
      N(60, 300, 'A8 · Home, first run (part 1)', { btn: 1 }), N(300, 300, 'B1 · Home', { btn: 1 }),
      N(540, 40, 'B2 · Log a spend, empty', { btn: 1, dim: 1, sheetH: '130px' }), N(780, 40, 'B3 · Log a spend, filled', { btn: 1, dim: 1, sheetH: '150px' }), N(1020, 40, 'B4 · Date picker', { btn: 1, dim: 1, sheetH: '160px' }),
      N(780, 300, 'B5 · Push permission', { btn: 1, dim: 1, alert: 1 }),
      N(540, 560, 'B6 · Spend list, month switcher', { cardTop: '60px', cardBottom: '8px' }), N(780, 560, 'B6b · Category detail', { cardTop: '60px', cardBottom: '8px' }), N(1020, 560, 'B7 · Spend detail', { cardTop: '60px', cardBottom: '8px', dim: 1, sheetH: '150px' }), N(1260, 560, 'B8 · Delete confirm', { cardTop: '60px', cardBottom: '8px', dim: 1, sheetH: '150px', alert: 1 }),
      N(60, 560, 'B1 · Home, offline', { btn: 0.4, banner: 'grey' }),
    ];
    const flowLabels = [
      L(227, 404, 'returning member'), L(490, 144, 'Log a spend'), L(707, 144, 'type an amount'), L(947, 144, 'change the date'),
      L(958, 305, 'first ever spend'), L(560, 330, 'Save · numbers animate up'), L(588, 440, 'Allow or Don\\'t Allow'),
      L(462, 664, 'tap the hero card · or the month chip'), L(707, 664, 'tap a Category'), L(947, 664, 'tap a spend'), L(1187, 664, 'Delete'), L(1068, 872, 'confirmed'),
      L(250, 245, 'Settings chip'), L(250, 780, 'See last month'), L(420, 230, 'preview card'), L(215, 560, 'no connection'), L(620, 556, 'pick a month'),
    ];
    return { ...common, catRows, flowNodes, flowLabels, catsSnacks: mkCats('snacks'), catsClothes: mkCats('clothes'),
      capMinor: 99999999999900,
      n80: 80, n900: 900, n1190: 1190, n2000: 2000, n18000: 18000 };
  }`,
  body: page('My Money Recap · Home and the daily loop (revised)',
    'Part 2, redrawn against SCREEN-STATES.md. Home states that a decision changed (D41, D54, D47) and the ones that were never drawn: no target, the before-payday warning (D53), a partial month in the list, offline (D58/D63), the first-load skeleton (D49), the 12-digit cap (D57), the month switcher (D59) and the read-only date on an old spend. Ledger: Maya, GBP, salary £2,400, payday 25, target 20% = £480, Otter; 26 September unless the sticky says otherwise. Yellow stickies are annotations, not UI.',
    [row('Row 1 · Home, states a decision changed or added', [b1Before, b1NoSalary, b1NoTarget, b1WarnBefore, b1Partial]),
     row('Row 2 · Offline, loading, the preview card and the amount cap', [b1Offline, b1FirstLoad, b1Preview, b2TooLong]),
     row('Row 3 · The Spend list', [b6, b6Open, b6Empty, b7Old]),
     row('Row 4 · Flow', [flowB])]),
};
