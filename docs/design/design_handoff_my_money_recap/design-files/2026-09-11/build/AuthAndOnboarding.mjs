import { X, col, row, page, screen, backRow, homeButtons, previewCard, targetCardRow, flow, logicCommon } from './lib.mjs';

const appleBtn = (dim = false) => `<button${dim ? ' disabled=""' : ''} style="display:flex;align-items:center;justify-content:center;gap:10px;width:100%;height:52px;border:0;border-radius:999px;background:#000;color:#fff;font:600 17px/1 -apple-system,'SF Pro Text',system-ui,sans-serif;cursor:pointer${dim ? ';opacity:.4' : ''}"><svg width="17" height="20" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"><path d="M12.152 6.896c-.948 0-2.415-1.078-3.96-1.04-2.04.027-3.91 1.183-4.961 3.014-2.117 3.675-.546 9.103 1.519 12.09 1.013 1.454 2.208 3.09 3.792 3.039 1.52-.065 2.09-.987 3.935-.987 1.831 0 2.35.987 3.96.948 1.637-.026 2.676-1.48 3.676-2.948 1.156-1.688 1.636-3.325 1.662-3.415-.039-.013-3.182-1.221-3.22-4.857-.026-3.04 2.48-4.494 2.597-4.559-1.429-2.09-3.623-2.324-4.39-2.376-2-.156-3.675 1.09-4.61 1.09zM15.53 3.83c.843-1.012 1.4-2.427 1.245-3.83-1.207.052-2.662.805-3.532 1.818-.78.896-1.454 2.338-1.273 3.714 1.338.104 2.715-.688 3.559-1.701"></path></svg>Sign in with Apple</button>`;
const googleBtn = `<button style="display:flex;align-items:center;justify-content:center;gap:12px;width:100%;height:52px;border:2px solid var(--mmr-palette-ink-100);box-sizing:border-box;border-radius:999px;background:var(--mmr-palette-white);color:var(--mmr-palette-ink-900);font:800 16px/1 var(--mmr-font-family-text);cursor:pointer"><svg width="20" height="20" viewBox="0 0 48 48" aria-hidden="true"><path fill="#EA4335" d="M24 9.5c3.54 0 6.71 1.22 9.21 3.6l6.85-6.85C35.9 2.38 30.47 0 24 0 14.62 0 6.51 5.38 2.56 13.22l7.98 6.19C12.43 13.72 17.74 9.5 24 9.5z"></path><path fill="#4285F4" d="M46.98 24.55c0-1.57-.15-3.09-.38-4.55H24v9.02h12.94c-.58 2.96-2.26 5.48-4.78 7.18l7.73 6c4.51-4.18 7.09-10.36 7.09-17.65z"></path><path fill="#FBBC05" d="M10.53 28.59c-.48-1.45-.76-2.99-.76-4.59s.27-3.14.76-4.59l-7.98-6.19C.92 16.46 0 20.12 0 24c0 3.88.92 7.54 2.56 10.78l7.97-6.19z"></path><path fill="#34A853" d="M24 48c6.48 0 11.93-2.13 15.89-5.81l-7.73-6c-2.15 1.45-4.92 2.3-8.16 2.3-6.26 0-11.57-4.22-13.47-9.91l-7.98 6.19C6.51 42.62 14.62 48 24 48z"></path></svg>Continue with Google</button>`;
const faceId = `<svg width="64" height="64" viewBox="0 0 64 64" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M4 20V10a6 6 0 0 1 6-6h10M44 4h10a6 6 0 0 1 6 6v10M60 44v10a6 6 0 0 1-6 6H44M20 60H10a6 6 0 0 1-6-6V44"></path><path d="M22 26v4M42 26v4M32 26v10h-4"></path><path d="M22 42c3 3 6 4 10 4s7-1 10-4"></path></svg>`;
const emailInput = (value) => `<input type="email" inputmode="email" autocomplete="email" value="${value}" style="box-sizing:border-box;width:100%;height:52px;border:2px solid var(--mmr-palette-ink-100);border-radius:999px;background:var(--mmr-palette-ink-100);padding:0 20px;font:700 16px var(--mmr-font-family-text);color:var(--mmr-color-text-primary);outline:none">`;
const a2With = (label, errorText, sendDisabled, noteText) => col(label, screen('recap', `${backRow('Email')}
  <div style="margin-top:18px">${X('Headline', 'size="m"', 'Where should we send your link?', '100%,84px')}</div>
  <div style="margin-top:24px">${X('Card', '', `<label style="display:flex;flex-direction:column;gap:8px"><span style="font:800 13px/1.2 var(--mmr-font-family-text);color:var(--mmr-color-text-secondary)">Email address</span>${emailInput('maya@example.com')}</label>
    ${X('InlineError', '', errorText, '100%,44px')}`, '100%,180px')}</div>
  <div style="margin-top:auto">${X('Button', sendDisabled ? 'state="disabled"' : '', 'Send me a link', '100%,52px')}</div>`), [noteText]);

// Row 1 · launch and sign-in states
const a0b = col('A0b · Unlock', screen('recap', `<div style="flex-grow:1;display:flex;flex-direction:column;align-items:center;justify-content:center;gap:28px">
    <img src="{{ coverSrc }}" alt="" style="width:270px;height:auto;display:block">
    <div style="font: 700 32px/1.1 var(--mmr-font-family-display); letter-spacing: -.01em; text-align: center">My Money Recap</div>
  </div>
  <div style="display:flex;flex-direction:column;align-items:center;gap:10px;padding-bottom:12px;color:var(--mmr-palette-white)">${faceId}<div style="font:800 14px/1.2 var(--mmr-font-family-text)">Face ID</div></div>`), [
  `D61 / D62: a session saved within the last 30 days exists, so the cover asks for Face ID (or Touch ID) and goes straight to Home. Works with no network — the unlock releases the Keychain entry and Home renders from cache. Each successful unlock extends the session another 30 days from last use.`,
  'The Face ID glyph and sheet are iOS-owned (LAContext); drawn for placement only, on the launch screen so nothing flashes between launch and Home. Three failures → A1 Sign in.']);
const a0bPasscode = col('A0b · Unlock / passcode fallback', screen('recap', `<div style="flex-grow:1;display:flex;flex-direction:column;align-items:center;justify-content:center;gap:28px">
    <img src="{{ coverSrc }}" alt="" style="width:270px;height:auto;display:block">
    <div style="font: 700 32px/1.1 var(--mmr-font-family-display); letter-spacing: -.01em; text-align: center">My Money Recap</div>
  </div>
  <div style="display:flex;flex-direction:column;gap:14px">
    ${X('SubCopy', 'dc-props="{{ centred }}"', 'Enter your passcode to open My Money Recap.', '100%,44px')}
    ${X('Button', 'variant="secondary"', 'Enter passcode', '100%,52px')}
  </div>`), [
  'Biometrics unavailable or not enrolled: the fallback line and a secondary button that hands off to the device passcode sheet. Same three-failure rule.']);
const a1Expired = col('A1 · Sign in / session expired', screen('recap', `${X('Headline', 'size="l"', 'Your month in money.', '100%,98px')}
  <div style="margin-top:10px">${X('SubCopy', '', 'Sign in to start your recap.', '100%,22px')}</div>
  ${X('Illustration', 'src="{{ coverSrc }}" width="{{ n230 }}"', '', '100%,190px')}
  ${X('Card', 'under-illustration="{{ true }}"', `<div style="display:flex;flex-direction:column;gap:12px">
      ${X('HelperText', '', "It's been a while — sign in again to pick up where you left off.", '100%,40px')}
      ${appleBtn()}
      ${googleBtn}
      ${X('Button', '', 'Continue with email', '100%,52px')}
    </div>`, '100%,300px')}
  <div style="margin-top:14px;font:700 13px/1.4 var(--mmr-font-family-text);color:var(--mmr-color-text-on-ground-muted);text-align:center;text-wrap:pretty">By continuing you agree to the Terms and Privacy Policy.</div>`), [
  `D62: after 30 days of no use the Keychain entry and the cache are erased and the next launch shows Sign in with one line above the buttons. The cover illustration drops from 260 to 230pt to give the card the extra line without moving the buttons off screen.`]);
const a3Expired = a2With('A3 · Check your inbox / link expired', 'That link has expired. Links work for 15 minutes.', false, `Tapping a link older than 15 minutes returns the Member to A2 with the address prefilled and an InlineError naming the cause. Send is enabled: one tap sends a fresh link.`);
const a3Used = a2With('A3 · Check your inbox / link already used', 'That link has already been used. We can send another.', false, `Links are single use. Same return to A2, Send enabled.`);
const a3Rate = a2With('A3 · Check your inbox / rate limited', "That's five links in an hour. Try again later.", true, `Five links per email per hour (§11). Send is disabled at 40% until the window passes; the address stays so nothing has to be retyped.`);

// Row 2 · onboarding
const a5Empty = col('A5 · Currency / empty search', screen('recap', `${backRow('Step 2 of 4')}
  <div style="margin-top:18px">${X('Headline', 'size="m"', 'Which currency?', '100%,42px')}</div>
  <div style="margin-top:20px">${X('Card', '', `<div style="display:flex;align-items:center;gap:10px;height:52px;box-sizing:border-box;border:2px solid var(--mmr-palette-ink-100);border-radius:999px;background:var(--mmr-palette-ink-100);padding:0 18px">
      <svg width="18" height="18" viewBox="0 0 20 20" fill="none" stroke="var(--mmr-palette-ink-500)" stroke-width="2.2" stroke-linecap="round" aria-hidden="true"><circle cx="8.5" cy="8.5" r="6"></circle><path d="M13 13l5 5"></path></svg>
      <input type="search" value="ducat" style="flex-grow:1;border:0;background:transparent;font:700 16px var(--mmr-font-family-text);color:var(--mmr-color-text-primary);outline:none;padding:0;min-width:0">
    </div>
    <div style="display:flex;flex-direction:column;align-items:center;gap:10px;padding:48px 0;text-align:center">
      ${X('HelperText', '', 'No currency matches "ducat".', '100%,20px')}
      <button style="height:44px;padding:0 4px;border:0;background:transparent;color:var(--mmr-palette-violet-500);font:800 14px var(--mmr-font-family-text);text-decoration:underline;text-underline-offset:3px;cursor:pointer">Clear search</button>
    </div>`, '100%,220px')}</div>
  <div style="margin-top:auto">${X('Button', '', 'Next', '100%,52px')}</div>`), [
  `Search matches on ISO code or name; nothing matches "ducat". One line and a link that clears the field. GBP stays selected underneath, so Next is still enabled.`]);
const amountCard = (valueText, cls = '', extra = '') => `<div style="margin-top:20px">${X('Card', '', `<span style="font:800 13px/1.2 var(--mmr-font-family-text);color:var(--mmr-color-text-secondary)">Amount, after tax</span>
    <div style="display:flex;align-items:baseline;gap:8px;border-bottom:2px solid ${extra ? 'var(--mmr-palette-coral-500)' : 'var(--mmr-palette-ink-100)'};padding-bottom:10px;margin-top:-4px">
      <span style="font:700 ${cls ? '22px' : '32px'}/1 var(--mmr-font-family-display);color:var(--mmr-color-text-secondary)">£</span>
      <span style="font:700 ${cls || '48px'}/1 var(--mmr-font-family-display);letter-spacing:-.02em;color:var(--mmr-color-text-primary);font-variant-numeric:tabular-nums">${valueText}</span>
      <span style="width:2px;height:${cls ? '26px' : '40px'};background:var(--mmr-palette-violet-500);border-radius:999px;align-self:center"></span>
    </div>${extra}`, extra ? '100%,160px' : '100%,130px')}</div>`;
const paydayCard = (valueAttr, nextKey) => `<div style="margin-top:12px">${X('Card', '', `${X('CardTitle', '', 'Payday', '100%,22px')}${X('DayPicker', `value="${valueAttr}" next-paydays="{{ ${nextKey} }}"`, '', '100%,80px')}`, '100%,150px')}</div>`;
const a6 = col('A6 · Salary and payday', screen('recap', `${backRow('Step 3 of 4')}
  <div style="margin-top:18px">${X('Headline', 'size="m"', 'Your monthly salary', '100%,42px')}</div>
  ${amountCard('2,400')}
  ${paydayCard('{{ n25 }}', 'next25')}
  <div style="margin-top:auto;display:flex;flex-direction:column;gap:4px">${X('Button', '', 'Next', '100%,52px')}<button style="height:44px;border:0;background:transparent;color:var(--mmr-palette-white);font:800 16px var(--mmr-font-family-text);cursor:pointer">Skip for now</button></div>`), [
  `D51: the shipped DayPicker — 1–31 plus "Last day" after 31, nothing disabled, the selection as the ink pill scrolled to centre — and the computed helper "Next paydays: 25 Sep · 25 Oct · 25 Nov", updating on every tap. There is no "If the month is shorter…" helper line; the three dates say it (tap 31 and read "30 Sep · 31 Oct · 30 Nov").`,
  'The dates are a prop: min(Payday, days in month) is computed by the screen, not the picker (§8.2).']);
const a6TooLong = col('A6 · Salary and payday / amount too long', screen('recap', `${backRow('Step 3 of 4')}
  <div style="margin-top:18px">${X('Headline', 'size="m"', 'Your monthly salary', '100%,42px')}</div>
  ${amountCard('999,999,999,999', '28px', `<span style="font:700 13px/1.4 var(--mmr-font-family-text);color:var(--mmr-color-text-accent-coral)">That's more than the app can store.</span>`)}
  ${paydayCard('{{ n25 }}', 'next25')}
  <div style="margin-top:auto;display:flex;flex-direction:column;gap:4px">${X('Button', '', 'Next', '100%,52px')}<button style="height:44px;border:0;background:transparent;color:var(--mmr-palette-white);font:800 16px var(--mmr-font-family-text);cursor:pointer">Skip for now</button></div>`), [
  `D57: at 12 digits before the separator the field refuses further keystrokes and shows "That's more than the app can store." The string is wider than the field at 48 and 36pt, so AmountField draws it at 28pt (measured, §11, D71). The amount is still valid, so Next stays enabled.`]);
const a7Skipped = col('A7 · Saving target / salary skipped', screen('recap', `${backRow('Step 4 of 4')}
  <div style="margin-top:18px">${X('Headline', 'size="m"', 'How much do you want to keep?', '100%,84px')}</div>
  <div style="margin-top:20px">${X('PromptCard', 'title="Set a salary first" helper="A target is a share of your salary, so we need the salary first."', '', '100%,110px')}</div>
  <div style="margin-top:auto">${X('Button', '', 'Skip for now', '100%,52px')}</div>`), [
  `D44: a target cannot be set without a Salary. The slider card is replaced by a PromptCard; tapping it returns to step 3. The primary button reads "Skip for now" and goes to Home with both prompts (A8 · skipped). Replaces the drawn behaviour where a percentage could be set with no salary.`]);

// Row 3 · first-run Home
const firstRunHeader = `<div style="display:flex;justify-content:space-between;align-items:center">${X('Chip', 'as-label="{{ true }}"', 'September 2026', 'auto,30px')}${X('Chip', 'variant="inverse"', 'Settings', 'auto,30px')}</div>
  <div style="display:flex;justify-content:space-between;align-items:center;margin-top:18px">${X('Headline', 'size="m"', 'Hi, Maya', 'auto,42px')}${X('Avatar', 'src="{{ avatarSrc }}" label="Otter" size="{{ n56 }}"', '', '56px,56px')}</div>`;
const a8First = col('A8 · Home, first run', screen('recap', `${firstRunHeader}
  <div style="margin-top:18px">${X('Card', '', `${X('StatHero', 'eyebrow="Spent so far · salary lands in 21 days" value="{{ n0 }}" format="{{ gbp }}"', '', '100%,100px')}${X('StatChipRow', '', X('StatChip', 'tone="earned"', 'Money in £0', 'auto,34px'), '100%,34px')}`, '100%,191px')}</div>
  <div style="margin-top:12px">${X('Card', '', `${targetCardRow('Saving target', '£0 of £480')}${X('ProgressBar', 'value="{{ n0 }}" max="{{ n480 }}" tone="green"', '', '100%,22px')}${X('HelperText', '', 'Nothing spent yet. Streak day 1.', '100%,20px')}`, '100%,136px')}</div>
  <div style="margin-top:50px;display:flex;flex-direction:column;gap:12px">${X('Button', '', 'Log a spend', '100%,52px')}${X('Button', 'variant="secondary" state="disabled"', 'See last month', '100%,52px')}</div>
  ${previewCard('Your first recap arrives on 1 October.', 'Recaps')}`), [
  `D43: "Streak day 1" for a Member who signed up today (4 September) — the streak window is max(1st, sign-up date) to today. D41: no Saved chip while Money In is £0. D47: the recap card is tappable.`]);
const a8Skipped = col('A8 · Home / skipped', screen('recap', `${firstRunHeader}
  <div style="margin-top:18px">${X('Card', '', `${X('StatHero', 'eyebrow="Spent so far" value="{{ n0 }}" format="{{ gbp }}"', '', '100%,100px')}${X('StatChipRow', '', X('StatChip', 'tone="earned"', 'Money in £0', 'auto,34px'), '100%,34px')}`, '100%,191px')}</div>
  <div style="margin-top:12px">${X('PromptCard', 'size="hero" eyebrow="Salary" title="Set your monthly salary" helper="We\'ll show what you kept once we know what came in."', '', '100%,191px')}</div>
  <div style="margin-top:12px">${X('PromptCard', 'size="target" title="Set a saving target" helper="Pick a share of your salary to keep each month."', '', '100%,136px')}</div>
  <div style="margin-top:12px;display:flex;flex-direction:column;gap:12px">${X('Button', '', 'Log a spend', '100%,52px')}${X('Button', 'variant="secondary" state="disabled"', 'See last month', '100%,52px')}</div>
  ${previewCard('Your first recap arrives on 1 October.', 'Recaps')}`), [
  `D54: the hero card is no longer replaced wholesale — it keeps the StatHero (Spent £0, Money in £0) and the salary PromptCard moves below it. The target PromptCard is unchanged. Both prompts are the shipped component at and.`,
  'Home scrolls in this state (hero 191 + 191 + 136 before the buttons); "Log a spend" is above the fold, "See last month" and the recap card below it. See the sticky on B1 · no salary for the alternative.']);

const a3Web = col('A3 · Web fallback page <span style="color:var(--mmr-palette-ink-500)">· 1200 × 720, shown at 50%</span>', `<div style="width:600px;height:360px;overflow:hidden;flex:none;border-radius:16px"><div style="width:1200px;height:720px;transform:scale(.5);transform-origin:top left;box-sizing:border-box;background:var(--mmr-color-ground-recap);color:var(--mmr-color-text-on-ground);display:flex;flex-direction:column;align-items:center;justify-content:center;gap:40px;padding:80px;font-family:var(--mmr-font-family-text)">
  <img src="{{ envelopeSrc }}" alt="" style="width:300px;display:block">
  <div style="font:700 56px/1.05 var(--mmr-font-family-display);letter-spacing:-.01em;text-align:center;max-width:820px;text-wrap:balance">Open My Money Recap on your phone to finish signing in.</div>
  <div style="font:700 22px/1.4 var(--mmr-font-family-text);opacity:.85;text-align:center">The link in your email only works in the app.</div>
</div></div>`, [
  `Not an app screen. The magic-link URL doubles as this page when the tap happens on a device without the app (a laptop, D32). Violet ground, the envelope, one line, no controls — there is nothing to do here but pick up the phone.`], 600);

const flowA = flow('Flow · Launch to Home', 1980, 760, [
  'M156 354H278', 'M376 290C440 290 440 144 498 144', 'M376 400H938', 'M328 458C328 600 1868 600 1868 460',
  'M596 144H718', 'M816 144C880 144 880 320 938 320', 'M768 40C768 4 1868 4 1868 248',
  'M1036 354H1158', 'M1256 354H1378', 'M1476 354H1598', 'M1696 354H1818',
  'M108 458V488', 'M156 600C600 640 1600 700 1868 462', 'M156 560C230 560 230 430 278 430',
  'M1600 300C1560 260 1520 260 1478 300',
], 'arwA', 'Launch checks for a saved session first (D61/D62): under 30 days old it unlocks with Face ID or the passcode and goes straight to Home, with no network round-trip; three failures fall back to Sign in. Onboarding progress is saved per screen; every step after the first has a back chevron and re-selects the saved answer on resume. Step 4 is blocked when the salary was skipped (D44): its prompt returns to step 3 and "Skip for now" goes to Home. Thumbnails are schematic; see rows 1–3 for the screens.', 10, 17, '', 'left:60px;top:28px;width:420px');

export default {
  props: { showAnnotations: { editor: 'boolean', default: true, tsType: 'boolean', section: 'Canvas' } },
  logic: `  renderVals() {${logicCommon}
    const pos = [[60, 250, 'A0 · Launch', '38%', '38%', 0], [280, 250, 'A1 · Sign in', '48%', '8px', 1], [500, 40, 'A2 · Enter email', '30%', '40%', 1], [720, 40, 'A3 · Check your inbox', '45%', '40%', 0], [940, 250, 'A4 · Name and avatar', '30%', '16%', 1], [1160, 250, 'A5 · Currency', '30%', '16%', 1], [1380, 250, 'A6 · Salary and payday', '30%', '20%', 1], [1600, 250, 'A7 · Saving target', '30%', '30%', 1], [1820, 250, 'A8 · Home', '28%', '30%', 1], [60, 490, 'A0b · Unlock (Face ID)', '38%', '38%', 0]];
    const flowNodes = pos.map(([x, y, label, cardTop, cardBottom, btn]) => N(x, y, label, { cardTop, cardBottom, btn }));
    const mut = 'var(--mmr-palette-ink-500)';
    const flowLabels = [L(218, 340, 'no session · auto'), L(420, 200, 'Continue with email'), L(658, 384, 'Apple or Google, first sign-in'), L(1098, 592, 'Apple or Google, returning member'), L(658, 130, 'Send me a link'), L(900, 230, 'tap link, first sign-in'), L(1318, 20, 'tap link, returning member'), L(1097, 340, 'Next'), L(1317, 340, 'Next'), L(1537, 340, 'Next'), L(1537, 372, 'Skip for now → A7 blocked', mut), L(1757, 340, 'Done'), L(1757, 372, 'Skip for now', mut), L(225, 473, 'saved session, under 30 days'), L(1000, 668, 'Face ID or passcode · no network needed'), L(232, 530, '3 failures', mut), L(1540, 268, 'Set a salary first', mut)];
    return { ...common, flowNodes, flowLabels, n25: 25, n230: 230, centred: { style: { textAlign: 'center' } }, next25: ['25 Sep', '25 Oct', '25 Nov'] };
  }`,
  body: page('My Money Recap · Auth and onboarding',
    'Part 1. The biometric unlock (D61/D62), the expired-session sign-in, the three magic-link failures and the web fallback page, the empty currency search, step 3 with the DayPicker (D51), the 12-digit cap, the blocked step 4 (D44) and the two first-run Homes (D43, D54). Ledger: Maya, GBP, salary £2,400, payday 25, target 20% = £480, Otter; signed up 4 September. Yellow stickies are annotations, not UI.',
    [row('Row 1 · Launch, unlock and sign-in states', [a0b, a0bPasscode, a1Expired, a3Expired, a3Used, a3Rate]),
     row('Row 2 · Onboarding', [a5Empty, a6, a6TooLong, a7Skipped, a3Web]),
     row('Row 3 · First-run Home', [a8First, a8Skipped]),
     row('Row 4 · Flow', [flowA])]),
};
