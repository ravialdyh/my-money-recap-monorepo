import { X, col, row, page, screen, sheetBack, flow, logicCommon, NEW, REDRAW } from './lib.mjs';

const head = (title) => `<div style="display:flex;flex-direction:column;gap:4px">${sheetBack}${X('Headline', 'size="m"', title, '100%,42px')}</div>`;
const full = (inner) => X('Sheet', 'variant="full" dc-props="{{ fullSheet }}" style="flex-grow:1;display:flex;flex-direction:column"', inner, '100%,100%');
const section = (eyebrow, rows, h) => `<div style="display:flex;flex-direction:column;gap:6px">${X('Eyebrow', '', eyebrow, 'auto,16px')}${X('DataRowList', '', rows.join(''), `100%,${h}px`)}</div>`;
const LR = (label, value, extra = '') => X('ListRow', `label="${label}"${value ? ` value="${value}"` : ''} ${extra}`, '', '100%,56px');
const dim = (inner) => `<div style="opacity:.4">${inner}</div>`;

const d1Push = col('D1 · Settings / push denied', screen('recap', full(`${head('Settings')}
    ${section('You', [LR('Profile', 'Maya'), LR('Currency', 'GBP')], 112)}
    ${section('Money', [LR('Salary and payday', '£2,400 · 25th'), LR('Saving target', '20%')], 112)}
    ${section('Alerts', [LR('Notifications', 'Off in iOS Settings')], 56)}
    ${section('Your data', [LR('Export spends', ''), LR('Delete account', '', 'tone="destructive" trailing="none"'), LR('Log out', '', 'trailing="none"')], 168)}`)), [
  `${NEW}Push was denied at the system prompt, so the Notifications row reads "Off in iOS Settings" instead of a count. The preferences underneath are kept; the row still opens D8, whose toggles deep-link to iOS Settings.`]);
const d1Offline = col('D1 · Settings / offline', screen('recap', full(`${head('Settings')}
    ${X('Banner', 'tone="offline"', 'No connection. You can look, but not log.', '100%,60px')}
    ${dim(section('You', [LR('Profile', 'Maya'), LR('Currency', 'GBP')], 112))}
    ${dim(section('Money', [LR('Salary and payday', '£2,400 · 25th'), LR('Saving target', '20%')], 112))}
    ${dim(section('Alerts', [LR('Notifications', '2 on')], 56))}
    <div style="display:flex;flex-direction:column;gap:6px">${X('Eyebrow', '', 'Your data', 'auto,16px')}${X('DataRowList', '', `${X('ListRow', 'label="Export spends" state="disabled"', '', '100%,56px')}${X('ListRow', 'label="Delete account" tone="destructive" trailing="none" state="disabled"', '', '100%,56px')}${LR('Log out', '', 'trailing="none"')}`, '100%,168px')}</div>`)), [
  `${NEW}§6.19: rows render from cache; every row that writes sits at 40% behind the offline Banner. Log out stays live — it erases the Keychain entry and the cache on device, and the end-session call can fail silently.`]);

const salaryCards = (extraLine) => `${X('Card', 'dc-props="{{ tintCard }}"', X('AmountField', 'value-minor="{{ n240000 }}" currency="GBP" label="Monthly salary" on-change="{{ noop }}"', '', '100%,72px'), '100%,110px')}
    ${X('Card', 'dc-props="{{ tintCard }}"', `${X('CardTitle', '', 'Payday', '100%,22px')}${X('DayPicker', 'value="{{ n25 }}" next-paydays="{{ next25 }}"', '', '100%,80px')}`, '100%,150px')}
    ${X('HelperText', '', 'Your new payday starts in October.', '100%,20px')}
    ${extraLine || ''}
    ${X('Button', '', 'Save', '100%,52px')}`;
const d3 = col('D3 · Salary and payday', screen('recap', full(`${head('Salary and payday')}${salaryCards()}`)), [
  `${REDRAW}D51: the shipped DayPicker with "Last day" and the three-date helper, as on A6. The three dates are the next three paydays counted from today (27 September), so the list starts at 25 October — the 25 September payday has already happened. D56: Payday changes are forward-only, no sheet, so the line under the payday card reads "Your new payday starts in October.", replacing the old info Banner; the salary asks its own question on Save (next artboard).`,
  'Editing the salary before this month\'s effective Payday saves with no sheet and applies to the current month; on or after it, Save opens "which salary is this".']);
const d3Which = col('D3 · Salary and payday / which salary is this', screen('recap', `${full(`${head('Salary and payday')}${salaryCards()}`)}
  <div style="position:absolute;inset:0">${X('Sheet', 'title="Which salary is this?" scrim="{{ true }}"', `<div style="display:flex;flex-direction:column;gap:16px">
      <div style="font:700 15px/1.45 var(--mmr-font-family-text);color:var(--mmr-color-text-secondary);text-wrap:pretty">You were paid on 25 September. A date alone can't tell a correction from a raise, so tell us which this is.</div>
      ${X('Card', 'dc-props="{{ choiceCard }}"', X('DataRowList', '', `${X('ListRow', 'label="{{ rowA }}" trailing="chevron"', '', '100%,64px')}${X('ListRow', 'label="{{ rowB }}" trailing="chevron"', '', '100%,64px')}`, '100%,128px'), '100%,136px')}
      ${X('Button', 'variant="secondary"', 'Cancel', '100%,52px')}
    </div>`, '100%,100%')}</div>`), [
  `${NEW}D56: shown only when the current month's Payday has already passed — here 27 September, editing £2,400 → £2,800. The first row makes September's Money In £2,800 and its target £560; the second leaves September at £2,400 and Home reads "£2,800 from your next payday, 25 October". Cancel returns without saving.`,
  'The two rows sit in a tinted card, as rows do everywhere else in the app, each with a one-line consequence beneath the label. The sheet body line and the two sub-lines are proposed copy; SCREEN-STATES fixes only the two row labels.']);
const d4 = col('D4 · Saving target', screen('recap', full(`${head('Saving target')}
    ${X('Card', 'dc-props="{{ tintCard }}"', `<div style="font:700 64px/1 var(--mmr-font-family-display);letter-spacing:-.01em">20%</div><div style="padding:6px 0">${X('Slider', 'value="{{ n20 }}" min="{{ n0 }}" max="{{ n100 }}" step="{{ n5 }}" on-change="{{ noop }}"', '', '100%,44px')}</div>${X('HelperText', '', '£480 of £2,400 each month', '100%,20px')}`, '100%,200px')}
    ${X('HelperText', '', 'Your new target starts in October.', '100%,20px')}
    ${X('Button', '', 'Save', '100%,52px')}`)), [
  `${REDRAW}D65: forward-only, no sheet, and the screen says so — "Your new target starts in October." replaces the info Banner. A target cannot be mistyped the way a salary can, so there is no correction case to ask about; raising it retroactively would turn a met month into a missed one.`]);

const currencyList = (selectedIdr = true, idrValue = '') => `${X('Card', 'dc-props="{{ tintCard }}"', `${X('TextField', 'placeholder="Search by name or code"', '', '100%,52px')}${X('DataRowList', '', `${X('ListRow', `label="GBP · British Pound" trailing="check"${selectedIdr ? '' : ' selected="{{ true }}"'}`, '', '100%,56px')}${X('ListRow', `label="IDR · Indonesian Rupiah" trailing="check"${selectedIdr ? ' selected="{{ true }}"' : ''}${idrValue ? ' value="{{ changedValue }}"' : ''}`, '', '100%,56px')}${X('ListRow', 'label="BHD · Bahraini Dinar" trailing="check"', '', '100%,56px')}${X('ListRow', 'label="USD · US Dollar" trailing="check"', '', '100%,56px')}${X('ListRow', 'label="EUR · Euro" trailing="check"', '', '100%,56px')}${X('ListRow', 'label="SGD · Singapore Dollar" trailing="check"', '', '100%,56px')}`, '100%,336px')}`, '100%,420px')}`;
const d6Failed = col('D6 · Currency confirm / rate fetch failed', screen('recap', `${full(`${head('Currency')}${currencyList(true)}`)}
  <div style="position:absolute;inset:0">${X('Sheet', 'title="Convert everything to IDR?" scrim="{{ true }}"', `<div style="display:flex;flex-direction:column;gap:16px">
      ${X('InlineError', 'action="{{ tryAgain }}"', "We couldn't get today's rate.", '100%,44px')}
      <div style="font:700 15px/1.45 var(--mmr-font-family-text);color:var(--mmr-color-text-secondary);text-wrap:pretty">All your amounts will be converted once, at the rate we use. <strong style="color:var(--mmr-color-text-primary)">This can't be undone.</strong></div>
      <div style="display:flex;flex-direction:column;gap:10px">${X('Button', '', 'Enter it myself', '100%,52px')}${X('Button', 'variant="secondary"', 'Cancel', '100%,52px')}</div>
    </div>`, '100%,100%')}</div>`), [
  `${NEW}§8.9 step 1: the fetch failed, so the sheet opens with an InlineError carrying "Try again" in its action slot, and the primary button becomes "Enter it myself", which goes to D7 (manual rate). No before/after rows — there is no rate to compute them from. Cancel closes without changing anything.`]);
const d6Nothing = col('D6 · Currency / nothing to convert', screen('recap', full(`${head('Currency')}${currencyList(true, 'Changed ✓')}`)), [
  `${NEW}No Salary and no Spends: there is nothing to convert, so the confirm sheet is skipped and the currency changes on the tap. The row confirms the way the export button does (D67) — a "Changed ✓" value beside the check, no toast — and the value clears when the list is next opened.`]);

const exportCard = `${X('Card', 'dc-props="{{ tintCard }}"', `${X('CardTitle', '', 'Email me a CSV of every spend', '100%,22px')}${X('HelperText', '', "It goes to maya@example.com. Every spend you've ever logged, with its date, category, amount and note.", '100%,40px')}`, '100%,130px')}`;
const d9Sent = col('D9 · Export spends / sent', screen('recap', full(`${head('Export spends')}${exportCard}${X('Button', 'state="disabled"', 'Sent ✓', '100%,52px')}`)), [
  `${REDRAW}D67: the ink toast is deleted. The button itself becomes "Sent ✓" and dims to 40% for 60 seconds, so a double tap cannot send twice; then it reads "Send it" again. Toast is not built.`]);
const d9NoAddress = col('D9 · Export spends / no deliverable address', screen('recap', full(`${head('Export spends')}
    ${X('Card', 'dc-props="{{ tintCard }}"', `${X('CardTitle', '', 'Email me a CSV of every spend', '100%,22px')}${X('HelperText', '', "Every spend you've ever logged, with its date, category, amount and note.", '100%,40px')}${X('InlineError', 'action="{{ profileLink }}"', "We can't reach maya@privaterelay.appleid.com.", '100%,44px')}`, '100%,190px')}
    ${X('Button', 'state="disabled"', 'Send it', '100%,52px')}`)), [
  `${NEW}An Apple private-relay address that has bounced. InlineError inside the card, above the action, with a link to Profile in its action slot; "Send it" is disabled until the account has an address that delivers.`]);
const d11b = col('D11b · Restore account', screen('recap', X('Sheet', 'variant="full" dc-props="{{ coralSheet }}" style="flex-grow:1;display:flex;flex-direction:column"', `<div style="display:flex;flex-direction:column;gap:4px">${X('Headline', 'size="m"', 'Welcome back', '100%,42px')}</div>
    <div style="font:800 16px/1.45 var(--mmr-font-family-text);text-wrap:pretty">You asked us to delete this account today at 09:12. Nothing has gone yet.</div>
    <div style="font:700 15px/1.45 var(--mmr-font-family-text);color:var(--mmr-color-text-secondary);text-wrap:pretty">Keep it and every spend, every recap and every setting stays exactly as it was. Continue and it all goes at 09:12 tomorrow, for good.</div>
    <div style="margin-top:auto;display:flex;flex-direction:column;gap:10px">${X('Button', '', 'Keep my account', '100%,52px')}${X('Button', 'variant="secondary"', 'Continue with deletion', '100%,52px')}</div>`, '100%,100%')), [
  `${NEW}D60: shown immediately after a sign-in when deletion was requested under 24 hours ago, before Home or onboarding. Sheet variant="full" on the coral-100 tint; "Keep my account" (ink) clears the flag and continues to Home; "Continue with deletion" (secondary) signs out again. No back chevron — the Member has to answer.`,
  'The two body lines are proposed copy: SCREEN-STATES asks for the deletion time and what is lost, and fixes the headline and the two button labels.']);

const flowD = flow('Flow · Settings and account', 1920, 1000, [
  'M156 484H298', 'M348 380C348 164 440 164 538 164', 'M396 400H677Q707 400 707 370V194Q707 164 737 164H778', 'M396 420H917Q947 420 947 390V194Q947 164 977 164H1018', 'M396 440H1157Q1187 440 1187 410V194Q1187 164 1217 164H1258',
  'M348 588C348 704 440 704 538 704', 'M396 500H677Q707 500 707 530V674Q707 704 737 704H778', 'M396 520H917Q947 520 947 550V674Q947 704 977 704H1018',
  'M1356 164H1498', 'M1596 164H1738', 'M1548 60C1548 26 1308 26 1308 58', 'M1116 704H1258', 'M1356 704H1498', 'M300 560C230 560 230 704 158 704',
  'M60 760C20 760 20 888 108 888H1620Q1650 888 1650 858V734Q1650 704 1620 704H1598',
  'M828 60C828 30 760 30 760 60', 'M1740 300C1740 360 1650 360 1596 300',
  'M1548 808C1548 880 1788 880 1788 810',
], 'arwD', 'Salary edits after this month\'s Payday ask "which salary is this" (D56); Payday and Saving Target edits are forward-only with no sheet (D56, D65). A currency change rewrites every stored amount once, in one transaction, and records the rate; with no salary and no spends there is nothing to convert and the list changes directly. Signing in inside the 24-hour deletion window offers to restore (D60). Offline, every row that writes is disabled. Thumbnails are schematic; see rows 1–2 for the screens.', 15, 13,
  `<div style="position:absolute;left:1740px;top:380px;width:96px;height:120px;box-sizing:border-box;border:2px dashed var(--mmr-palette-ink-100);border-radius:12px;display:flex;align-items:center;justify-content:center;text-align:center;padding:8px;font:800 11px/1.3 var(--mmr-font-family-text);color:var(--mmr-palette-ink-500)">D6 · nothing to convert: no sheet</div>`, 'left:60px;top:28px;width:440px');

export default {
  props: { showAnnotations: { editor: 'boolean', default: true, tsType: 'boolean', section: 'Canvas' } },
  logic: `  renderVals() {${logicCommon}
    const flowNodes = [
      N(60, 380, 'B1 · Home (part 2)', { btn: 1 }),
      N(300, 380, 'D1 · Settings', { sheet: true, cardTop: '50px', cardBottom: '20px' }),
      N(540, 60, 'D2 · Profile', { sheet: true, btn: 1 }), N(780, 60, 'D3 · Salary and payday', { sheet: true, btn: 1, dim: 1, sheetH: '90px' }), N(1020, 60, 'D4 · Saving target', { sheet: true, btn: 1 }),
      N(1260, 60, 'D5 · Currency', { sheet: true, cardTop: '50px', cardBottom: '20px' }), N(1500, 60, 'D6 · Currency confirm', { sheet: true, cardTop: '50px', cardBottom: '20px', dim: 1, sheetH: '110px' }), N(1740, 60, 'D7 · Manual rate', { sheet: true, cardTop: '50px', cardBottom: '20px', dim: 1, sheetH: '120px' }),
      N(540, 600, 'D8 · Notifications', { sheet: true, cardTop: '50px', cardBottom: '60px' }), N(780, 600, 'D9 · Export spends', { sheet: true, cardTop: '50px', cardBottom: '110px', btn: 1 }), N(1020, 600, 'D10 · Delete account', { sheet: true, cardTop: '50px', cardBottom: '110px', btn: 1, coral: true }), N(1260, 600, 'D11 · Delete, confirm', { sheet: true, cardTop: '50px', cardBottom: '110px', btn: 1, coral: true }),
      N(1500, 600, 'A1 · Sign in (part 1)', { cardTop: '100px', cardBottom: '40px', btn: 1 }),
      N(60, 600, 'D12 · Log out', { sheet: true, cardTop: '50px', cardBottom: '20px', dim: 1, sheetH: '70px' }),
      N(1740, 600, 'D11b · Restore account', { sheet: true, cardTop: '40px', cardBottom: '60px', btn: 1 }),
    ];
    const mut = 'var(--mmr-palette-ink-500)';
    const flowLabels = [
      L(227, 484, 'Settings chip'), L(560, 470, 'tap a row'),
      L(1427, 164, 'pick a currency'), L(1667, 164, 'rate unavailable'), L(1427, 26, 'Convert · every amount rewritten'),
      L(1187, 704, 'Delete my account'), L(1427, 704, 'signed out immediately'),
      L(230, 632, 'Log out'), L(860, 888, 'Log out · signed out'),
      L(794, 22, 'Save after payday → which salary is this (D56)', mut), L(1670, 350, 'rate fetch failed → Enter it myself', mut),
      L(1668, 878, 'sign in under 24 h after deleting → restore (D60)', mut), L(300, 340, 'offline: writes disabled, Banner on top', mut),
    ];
    return { ...common, flowNodes, flowLabels, n5: 5, n20: 20, n25: 25, n100: 100, n240000: 240000, next25: ['25 Oct', '25 Nov', '25 Dec'],
      coralSheet: { style: { height: 'auto', flexGrow: 1, background: 'var(--mmr-palette-coral-100)' } },
      choiceCard: { style: { background: 'var(--mmr-color-surface-field)', padding: '4px 18px', gap: 0 } },
      rowA: R.createElement('span', { style: { display: 'flex', flexDirection: 'column', gap: 2 } }, 'The £2,800 I was paid on 25 September', R.createElement('span', { style: { fontSize: 12, fontWeight: 700, color: 'var(--mmr-color-text-secondary)' } }, 'Fixes September')),
      rowB: R.createElement('span', { style: { display: 'flex', flexDirection: 'column', gap: 2 } }, 'My new salary, from 25 October', R.createElement('span', { style: { fontSize: 12, fontWeight: 700, color: 'var(--mmr-color-text-secondary)' } }, 'September stays at £2,400')),
      changedValue: R.createElement('span', { style: { whiteSpace: 'nowrap' } }, 'Changed ✓'),
      tryAgain: R.createElement('a', { href: '#try-again' }, 'Try again'),
      profileLink: R.createElement('a', { href: '#profile' }, 'Profile') };
  }`,
  body: page('My Money Recap · Settings and account (revised)',
    'Part 4, redrawn against SCREEN-STATES.md. Settings is a white Sheet variant="full" on the violet ground; sub-screens reuse the onboarding cards. New: push denied and offline on the Settings list, the "which salary is this" sheet (D56), the two currency failures, the undeliverable export address and the restore-account sheet (D60). Redrawn: salary and payday on the shipped DayPicker with the forward-only line, the forward-only saving target (D65) and the export confirmation without a toast (D67). Today is 27 September 2026 throughout, so "next month" is October. Yellow stickies are annotations, not UI.',
    [row('Row 1 · Settings list and money', [d1Push, d1Offline, d3, d3Which, d4]),
     row('Row 2 · Currency, export and the restore sheet', [d6Failed, d6Nothing, d9Sent, d9NoAddress, d11b]),
     row('Row 3 · Flow', [flowD])]),
};
