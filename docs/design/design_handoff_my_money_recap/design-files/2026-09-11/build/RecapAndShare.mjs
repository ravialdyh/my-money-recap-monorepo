import { X, col, row, page, screen, withClose, backRow, flow, logicCommon, NEW, REDRAW } from './lib.mjs';

const closeNote = 'D46: a 44pt close control top-left (`plus` rotated 45°) and tappable PageDots (`onSelect`). The control belongs to the pager, not the screen component — it is laid over the screen and the first chip row moves 34pt right to clear it.';

// Row 1 · the four screens, rebuilt on the shipped components, with the close control.
const c1 = col('C1 · Recap 1, cover', withClose(X('RecapCoverScreen', 'month="September 2026" child-name="Maya" total="{{ n2400 }}" saved="{{ n640 }}" spent="{{ n1760 }}"', '', '390px,844px')), [
  `${REDRAW}Unchanged apart from the close control. Money In £2,400 (Salary landed on the 25th), Saved = 2,400 − 1,760 = £640, Spent £1,760. Computed live on every open.`, closeNote]);
const c2 = col('C2 · Recap 2, kept and streak', withClose(X('RecapEarningScreen', 'month="September" money-in="{{ n2400 }}" saved="{{ n640 }}" target-amount="{{ n480 }}" streak-days="{{ n9 }}" streak-range="{{ septStreak }}" streak-caption="12 to 20 September. New badge unlocked." days-in-month="{{ n30 }}"', '', '390px,844px'), true), [
  `${REDRAW}RecapEarningScreen used directly. kept% = round(640 ÷ 480 × 100) = 133, clamped 0–999. Streak 9 (12–20 September, nothing dated inside), ≥ 7 → "New badge unlocked."`, 'Yellow ground: the close control is ink, like the rest of the text.']);
const c3 = col('C3 · Recap 3, target and spending', withClose(X('RecapGoalScreen', 'saved="{{ n640 }}" target="{{ n480 }}" helper="Target met. £160 to spare." sub-copy="You kept a third more than you set out to keep." spent-total="{{ n1760 }}" categories="{{ septCats }}" splurge="£420 on a flight home. Your biggest single buy."', '', '390px,844px')), [
  `${REDRAW}Close control only. Four bars plus the roll-up already match the shipped default (maxCategoryBars = 4): Bills 35, Rides 26, Snacks 14, Clothes 10, then "Gifts, Games, Toys, Other · 15%". "£160 to spare" = Saved − target. Splurge £420 on the 27th.`]);
const c4 = col('C4 · Recap 4, archetype', withClose(X('RecapPersonalityScreen', 'child-name="Maya" archetype="squirrel" saved="{{ n640 }}" spent="{{ n1760 }}"', '', '390px,844px')), [
  `${REDRAW}Close control only. Still no PageDots — this is the screenshotted screen. Squirrel: Saved £640 ≥ £480 and Streak 9 ≥ 7 (§8.6, first match wins).`]);
const c7empty = col('C7 · Past recaps / empty', screen('recap', `${backRow('Recaps')}
  <div style="margin-top:16px">${X('Headline', 'size="m"', 'Past recaps', '100%,42px')}</div>
  <div style="flex-grow:1;display:flex;flex-direction:column;align-items:center;justify-content:center;gap:14px;text-align:center">
    <div style="font:700 22px/1.25 var(--mmr-font-family-display);text-wrap:balance;max-width:280px">Your first recap arrives on 1 October.</div>
    <button style="height:44px;padding:0 4px;border:0;background:transparent;color:var(--mmr-palette-white);font:800 16px var(--mmr-font-family-text);text-decoration:underline;text-underline-offset:4px;cursor:pointer">See September so far</button>
  </div>`), [
  `${NEW}No month with data yet. One centred line and a link to the month-so-far preview (P1). The current month is never listed here; the preview is how it is reached until the 1st.`]);

// Row 2 · variants.
const c2NoTarget = col('C2 · Recap 2 / no target set', withClose(X('RecapEarningScreen', 'month="September" money-in="{{ n2400 }}" saved="{{ n640 }}" target-amount="{{ nullValue }}" streak-days="{{ n9 }}" streak-range="{{ septStreak }}" streak-caption="12 to 20 September. New badge unlocked." days-in-month="{{ n30 }}"', '', '390px,844px'), true), [
  `${REDRAW}targetAmount={null} on the shipped component — never undefined, which the default would replace. The headline drops its second sentence and the compare card becomes the component's own no-target branch. The hand-composed version and its "cannot be reached through props" note are superseded.`]);
const c3Missed = col('C3 · Recap 3 / target missed', withClose(X('RecapGoalScreen', 'saved="{{ n310 }}" target="{{ n480 }}" helper="£170 short of your target." sub-copy="Close, but the coat tipped it." spent-total="{{ n2090 }}" categories="{{ octCats }}" splurge="£560 on a winter coat. Your biggest single buy."', '', '390px,844px')), [
  `${REDRAW}Close control only. October: Saved £310 ÷ £480 = 65%; "£170 short" = target − Saved. Bills 29, Clothes 28, Snacks 16, Rides 14 of £2,090; roll-up "Games, Other · 13%". Splurge £560 = 26.8% of Spent.`]);
const c3NoSpends = col('C3 · Recap 3 / no spends', withClose(X('RecapGoalScreen', 'saved="{{ n2400 }}" target="{{ n480 }}" helper="Target met. £1,920 to spare." sub-copy="Not a single spend logged this month." spent-total="{{ n0 }}" categories="{{ emptyList }}" splurge="{{ nullValue }}"', '', '390px,844px')), [
  `${REDRAW}Close control, and splurge={null} on the shipped component — its own no-spends branch draws "Nothing spent this month" and hides the splurge. Saved = the whole £2,400; 2,400 ÷ 480 = 500%, the bar clamps at full.`]);
const c3NoTarget = col('C3 · Recap 3 / no target set', withClose(X('RecapGoalScreen', 'target="{{ nullValue }}" spent-total="{{ n1760 }}" categories="{{ septCats }}" splurge="£420 on a flight home. Your biggest single buy."', '', '390px,844px')), [
  `${NEW}Reachable whenever step 4 was skipped. Headline "Where £1,760 went", SubCopy dropped — there is no percentage to state. The saving-target card becomes PromptCard size="target" (136pt, the height of the card it replaces) and opens Settings › Saving target. Bars and Splurge unchanged.`,
  'Drawn through the shipped RecapGoalScreen with `target={null}`, the same way C2 takes `targetAmount={null}`.']);
const c4Variant = (label, archetype, name, saved, spent, noteText) => col(label, withClose(X('RecapPersonalityScreen', `child-name="Maya" archetype="${archetype}" saved="{{ ${saved} }}" spent="{{ ${spent} }}"`, '', '390px,844px')), [noteText]);
const c4Treat = c4Variant('C4 · Recap 4 / Treat-Yourself', 'treatYourself', 'Maya', 'n310', 'n2090', `${REDRAW}RecapPersonalityScreen with archetype="treatYourself". October: Saved £310 < £480 and the £560 coat is 26.8% of Spent, ≥ 25%.`);
const c4Hustler = c4Variant('C4 · Recap 4 / Hustler', 'hustler', 'Maya', 'n640', 'n1760', `${REDRAW}RecapPersonalityScreen with archetype="hustler". The component's maxHeight={300} keeps the taller bee clear of the pill row, so the hand-composed 290pt version is retired. Saved £640 ≥ £480 but Streak < 7.`);
const c4Slow = c4Variant('C4 · Recap 4 / Slow Burner', 'slowBurner', 'Maya', 'n310', 'n2090', `${REDRAW}RecapPersonalityScreen with archetype="slowBurner". Target missed with no single spend at 25% or more of Spent.`);
const c4NoArchetype = col('C4 · Recap 4 / no archetype, no target', withClose(screen('personality', `${X('ChipRow', '', X('Chip', 'as-label="{{ true }}"', 'Your money personality', 'auto,30px'), '100%,30px')}
  <div style="margin-top:16px">${X('PromptCard', 'size="hero" title="Set a saving target to get an archetype" helper="Your archetype comes from what you kept against what you meant to keep."', '', '100%,191px')}</div>
  <div style="flex-grow:1"></div>
  ${X('Card', 'dc-props="{{ gap14 }}"', X('CompareBars', 'rows="{{ septRows }}"', '', '100%,60px'), '100%,104px')}
  <div style="margin-top:12px;margin-left:18px;opacity:.4">${X('ArchetypeRow', '', `${X('ArchetypePill', 'archetype="squirrel"', '', 'auto,24px')}${X('ArchetypePill', 'archetype="hustler"', '', 'auto,24px')}${X('ArchetypePill', 'archetype="treatYourself"', '', 'auto,24px')}${X('ArchetypePill', 'archetype="slowBurner"', '', 'auto,24px')}`, '100%,24px')}</div>`)), [
  `${NEW}§6.14: Money In > 0 but no Saving Target, so no Archetype. PromptCard size="hero" takes the mascot's place and opens Settings › Saving target; the Share button is hidden; the four pills sit at 40%. Saved vs Spent still compares — both are computable without a target.`,
  'The "This month, Maya, you were" line and the archetype name are omitted rather than replaced: the PRD names only the prompt for this state, so no headline copy is invented. The prompt\'s helper is the one shipped with PromptCard\'s preview.']);
const c1Retired = col('C1 · Recap 1 / no salary (retired)', `<div style="position:relative;width:390px;height:844px"><div style="opacity:.55">${screen('recap', `${X('ChipRow', '', `${X('Chip', 'as-label="{{ true }}"', 'September 2026', 'auto,30px')}${X('Chip', 'as-label="{{ true }}" variant="highlight"', "Maya's recap", 'auto,30px')}`, '100%,30px')}
  <div style="margin-top:22px">${X('Headline', 'size="l"', 'Your September in money', '100%,98px')}</div>
  ${X('Illustration', 'src="{{ coverSrc }}" width="{{ n310 }}"', '', '100%,260px')}
  ${X('Card', 'under-illustration="{{ true }}"', `${X('StatHero', 'eyebrow="Money that came your way" value="{{ n0 }}" format="{{ gbp }}"', '', '100%,100px')}
    ${X('PromptCard', 'title="Set your monthly salary to get a recap" helper="Opens Settings"', '', '100%,90px')}`, '100%,230px')}
  <div style="margin-top:18px">${X('PageDots', 'count="{{ n4 }}" active="{{ n0 }}" on-select="{{ noop }}"', '', 'auto,6px')}</div>`)}</div>
  <div style="position:absolute;left:24px;top:400px;right:24px;display:flex;justify-content:center"><span style="background:var(--mmr-palette-ink-900);color:var(--mmr-palette-white);border-radius:999px;padding:10px 18px;font:800 13px/1.2 var(--mmr-font-family-text);letter-spacing:.08em;text-transform:uppercase">Retired · see P1–P4</span></div></div>`, [
  `${REDRAW}§8.11: a month with Money In 0 and at least one Spend is a partial month, so this case is P1–P4 (situation 3), not a variant of C1. Kept dimmed as the reference for the prompt copy — "Set your monthly salary to get a recap" — which P4's closing line carries as "Set your monthly salary to get a full recap."`]);

// Row 3 · share images for the other three Archetypes (1080 × 1920, shown at 50%).
const share = (label, mascotKey, name, title, month, saved, spent, noteText) => col(`${label} <span style="color:var(--mmr-palette-ink-500)">· 1080 × 1920, shown at 50%</span>`, `<div style="width:540px;height:960px;overflow:hidden;flex:none">
<div style="width:1080px;height:1920px;transform:scale(.5);transform-origin:top left;box-sizing:border-box;background:var(--mmr-color-ground-personality);color:var(--mmr-color-text-on-ground);display:flex;flex-direction:column;align-items:center;padding:140px 96px 96px;font-family:var(--mmr-font-family-text)">
  <img src="{{ ${mascotKey} }}" alt="" style="width:720px;max-height:620px;object-fit:contain;display:block">
  <div style="margin-top:24px;font:700 44px/1.2 var(--mmr-font-family-text);opacity:.9">This month, ${name}, you were</div>
  <div style="margin-top:4px;font:700 168px/1 var(--mmr-font-family-display);letter-spacing:-.01em;text-align:center">${title}</div>
  <div style="margin-top:28px;font:800 44px/1.2 var(--mmr-font-family-text);opacity:.9">${name} · ${month}</div>
  <div style="margin-top:72px;width:100%;background:var(--mmr-color-surface-card);color:var(--mmr-color-text-primary);border-radius:56px;padding:64px 72px;box-sizing:border-box;display:grid;grid-template-columns:1fr 1fr;gap:48px">
    <div style="display:flex;flex-direction:column;gap:12px">
      <div style="display:flex;align-items:center;gap:16px;font:800 30px/1.2 var(--mmr-font-family-text);letter-spacing:.06em;text-transform:uppercase;color:var(--mmr-color-text-secondary)"><span style="width:18px;height:18px;border-radius:999px;background:var(--mmr-color-data-saved);display:inline-block"></span>Saved</div>
      <div style="font:700 120px/1 var(--mmr-font-family-display)">${saved}</div>
    </div>
    <div style="display:flex;flex-direction:column;gap:12px">
      <div style="display:flex;align-items:center;gap:16px;font:800 30px/1.2 var(--mmr-font-family-text);letter-spacing:.06em;text-transform:uppercase;color:var(--mmr-color-text-secondary)"><span style="width:18px;height:18px;border-radius:999px;background:var(--mmr-color-data-spent);display:inline-block"></span>Spent</div>
      <div style="font:700 120px/1 var(--mmr-font-family-display)">${spent}</div>
    </div>
  </div>
  <div style="margin-top:auto;font:800 34px/1.2 var(--mmr-font-family-text);letter-spacing:.12em;text-transform:uppercase;opacity:.85">My Money Recap</div>
</div>
</div>`, [noteText, 'Same rule as C6: Archetype, month, Display Name, Saved and Spent only. Every mascot is capped at 620px tall inside its 720px slot (object-fit contain), so the four posters keep the name, the line and the card at the same y. That cap is what the Hustler needed.'], 540);
const c6Hustler = share('C6 · Share image / Hustler', 'hustlerSrc', 'Maya', 'The Hustler', 'September 2026', '£640', '£1,760', `${NEW}The tallest mascot. Without the 620px cap the bee pushes the stats card 180px down and the footer off the poster.`);
const c6Treat = share('C6 · Share image / Treat-Yourself', 'treatSrc', 'Maya', 'Treat-Yourself', 'October 2026', '£310', '£2,090', `${NEW}"Treat-Yourself" has no "The" — the title keeps the ARCHETYPE_TITLE form the component uses. It fits on one 168px line.`);
const c6Slow = share('C6 · Share image / Slow Burner', 'slowburnerSrc', 'Maya', 'The Slow Burner', 'October 2026', '£310', '£2,090', `${NEW}The widest title — "The Slow Burner" wraps to two lines at 168px, which is why the mascot cap matters here too.`);

// Row 4 · D66 accessibility-size demonstration: the four screens as one vertical scroll.
const dots = (i, light = false) => `<div style="margin-bottom:14px">${X('PageDots', `count="{{ n4 }}" active="{{ n${i} }}" on-select="{{ noop }}"${light ? ' on-light="{{ true }}"' : ''}`, '', 'auto,6px')}</div>`;
const a11y = col('Recap at accessibility text size', `<div class="mmr-a11y" style="width:390px;display:flex;flex-direction:column">
${X('Screen', 'ground="recap" dc-props="{{ autoScreen }}"', `${dots(0)}
  ${X('ChipRow', '', `${X('Chip', 'as-label="{{ true }}"', 'September 2026', 'auto,30px')}${X('Chip', 'as-label="{{ true }}" variant="highlight"', "Maya's recap", 'auto,30px')}`, '100%,30px')}
  <div style="margin-top:22px">${X('Headline', 'size="l"', 'Your September in money', '100%,130px')}</div>
  <div style="margin-top:20px">${X('Illustration', 'src="{{ coverSrc }}" width="{{ n240 }}"', '', '100%,200px')}</div>
  ${X('Card', 'under-illustration="{{ true }}"', `${X('StatHero', 'eyebrow="Money that came your way" value="{{ n2400 }}" format="{{ gbp }}"', '', '100%,100px')}
    ${X('StatChipRow', '', `${X('StatChip', 'tone="saved"', 'Saved £640', 'auto,34px')}${X('StatChip', 'tone="spent"', 'Spent £1,760', 'auto,34px')}`, '100%,34px')}
    ${X('HelperText', '', 'Scroll to see where it all went.', '100%,20px')}`, '100%,260px')}`, '390px,auto')}
${X('Screen', 'ground="earning" dc-props="{{ autoScreen }}"', `${dots(1, true)}
  ${X('ChipRow', '', `${X('Chip', 'as-label="{{ true }}" variant="inverse"', 'Kept', 'auto,30px')}${X('Chip', 'as-label="{{ true }}"', 'September', 'auto,30px')}`, '100%,30px')}
  <div style="margin-top:12px">${X('Headline', 'size="compact"', "You kept £640 of £2,400. That's 133% of your target.", '100%,200px')}</div>
  <div style="margin-top:12px">${X('Illustration', 'src="{{ jarSrc }}" width="{{ n160 }}" dc-props="{{ jarStyle }}"', '', '100%,160px')}</div>
  ${X('Card', 'dc-props="{{ gap12 }}"', `${X('Eyebrow', '', 'Kept against your target', 'auto,16px')}${X('CompareBars', 'scale="max" rows="{{ keptRows }}"', '', '100%,60px')}`, '100%,130px')}
  <div style="margin-top:12px">${X('Card', 'variant="accent"', `<div style="display:flex;align-items:center;gap:14px">${X('StreakBadge', 'value="{{ n9 }}"', '', '64px,64px')}<div style="display:grid;gap:2px">${X('CardTitle', '', 'Longest no-spend streak', 'auto,22px')}<div style="font-size:18px;font-weight:700;opacity:.8">12 to 20 September. New badge unlocked.</div></div></div>
    ${X('StreakDots', 'total="{{ n30 }}" active-range="{{ septStreak }}"', '', '100%,36px')}`, '100%,200px')}</div>`, '390px,auto')}
${X('Screen', 'ground="goal" dc-props="{{ autoScreen }}"', `${dots(2)}
  ${X('ChipRow', '', `${X('Chip', 'as-label="{{ true }}"', 'Target', 'auto,30px')}${X('Chip', 'as-label="{{ true }}" variant="inverse"', 'Spending', 'auto,30px')}`, '100%,30px')}
  <div style="margin-top:16px">${X('Headline', 'size="compact"', '133% of your saving target', '100%,100px')}</div>
  <div style="margin-top:8px">${X('SubCopy', '', 'You kept a third more than you set out to keep.', '100%,60px')}</div>
  <div style="margin-top:18px">${X('Card', '', `<div style="display:flex;flex-direction:column;gap:4px">${X('CardTitle', '', 'Saving target', 'auto,22px')}<span style="font:800 19px/1.2 var(--mmr-font-family-text);color:var(--mmr-color-text-secondary)">£640 of £480</span></div>
    <div style="padding-top:14px">${X('ProgressBar', 'value="{{ n640 }}" max="{{ n480 }}" marker="{{ piggy }}"', '', '100%,22px')}</div>
    ${X('HelperText', '', 'Target met. £160 to spare.', '100%,20px')}`, '100%,180px')}</div>
  <div style="margin-top:12px">${X('Card', '', `<div>${X('Eyebrow', '', 'Where £1,760 went', 'auto,16px')}${X('CardTitle', '', 'Mostly bills', '100%,22px')}</div>
    <div style="display:grid;gap:10px"><sc-for list="{{ septBars }}" as="b" hint-placeholder-count="4">${X('CategoryBar', 'label="{{ b.label }}" percent="{{ b.percent }}" tone="{{ b.tone }}"', '', '100%,48px')}</sc-for>${X('HelperText', '', 'Gifts, Games, Toys, Other · 15%', '100%,24px')}</div>
    <div style="display:grid;gap:2px">${X('Eyebrow', 'tone="coral"', 'The splurge', 'auto,16px')}<div style="font-size:20px;font-weight:800;line-height:1.3">£420 on a flight home. Your biggest single buy.</div></div>`, '100%,460px')}</div>`, '390px,auto')}
${X('Screen', 'ground="personality" dc-props="{{ autoScreen }}"', `${dots(3)}
  ${X('ChipRow', '', X('Chip', 'as-label="{{ true }}"', 'Your money personality', 'auto,30px'), '100%,30px')}
  <div style="margin-top:16px;font-size:20px;font-weight:800;opacity:.85">This month, Maya, you were</div>
  <div style="margin-top:2px">${X('Headline', 'size="display"', 'The Squirrel', '100%,180px')}</div>
  <div style="margin-top:12px;opacity:.9">${X('SubCopy', '', "You stashed more than you meant to, and you didn't crack once.", '100%,90px')}</div>
  <div style="margin-top:16px">${X('Illustration', 'src="{{ squirrelSrc }}" width="{{ n300 }}" max-height="{{ n260 }}"', '', '100%,260px')}</div>
  ${X('Card', 'under-illustration="{{ true }}" dc-props="{{ gap14 }}"', `${X('CompareBars', 'rows="{{ septRows }}"', '', '100%,60px')}${X('Button', 'icon="{{ shareIcon }}"', 'Share your Squirrel', '100%,64px')}`, '100%,220px')}
  <div style="margin-top:12px;margin-left:18px">${X('ArchetypeRow', '', `${X('ArchetypePill', 'archetype="squirrel" selected="{{ true }}"', '', 'auto,24px')}${X('ArchetypePill', 'archetype="hustler"', '', 'auto,24px')}${X('ArchetypePill', 'archetype="treatYourself"', '', 'auto,24px')}${X('ArchetypePill', 'archetype="slowBurner"', '', 'auto,24px')}`, '100%,24px')}</div>`, '390px,auto')}
</div>`, [
  `${NEW}D66, one demonstration artboard. At the "Large" accessibility size (about 1.35× the default ramp) the four screens become one vertical scroll; the PageDots at the top of each section are the section markers and tapping one scrolls to it. Nothing is clipped, nothing is capped: every card grows with its text. At default sizes nothing changes and the pager stays.`,
  'Type shown at 1.35×: headline l 64, m 54, compact 48, display 90, body 22, helper 19, buttons 22 in a 64pt pill. StatHero keeps its 80 → 56 → 40 rule on the formatted string. Illustrations are the one thing that shrinks, so the scroll stays about 3,400pt.']);

// Flow C, redrawn: close control, the partial-recap branch, the preview entry.
const flowC = flow('Flow · Recap and share', 1700, 1000, [
  'M156 380H298', 'M156 470C230 470 230 690 298 690', 'M396 690C480 690 480 470 398 470',
  'M396 404H538', 'M636 404H778', 'M876 404H1018', 'M1116 404H1258', 'M1356 404H1498',
  'M1068 300C1068 200 108 200 108 298',
  'M108 508C108 900 470 900 538 790', 'M396 740H538', 'M636 724H778', 'M876 724H1018', 'M1116 724H1258',
  'M1308 620C1308 560 148 560 148 508',
], 'arwC', 'Four horizontally paged screens, each with a 44pt close control top-left (D46) that returns to wherever the Recap was opened from. The month-so-far preview and the Partial Recap are the same four screens (D64), reached from the Home preview card before the first Recap and from a "Partial" row in the list after it. Thumbnails are schematic; see rows 1–3 for the screens.', 12, 16);

export default {
  props: { showAnnotations: { editor: 'boolean', default: true, tsType: 'boolean', section: 'Canvas' } },
  logic: `  renderVals() {${logicCommon}
    const septCats = [['bills', 35], ['rides', 26], ['snacks', 14], ['clothes', 10], ['gifts', 7], ['games', 5], ['toys', 2], ['other', 1]].map(([category, percent]) => ({ category, percent }));
    const octCats = [['bills', 29], ['clothes', 28], ['snacks', 16], ['rides', 14], ['games', 7], ['other', 6]].map(([category, percent]) => ({ category, percent }));
    const TONE = { bills: 'ink', rides: 'sky', snacks: 'coral', clothes: 'green', gifts: 'gold', games: 'violet', toys: 'yellow', other: 'rust' };
    const LABEL = { bills: 'Bills', rides: 'Rides', snacks: 'Snacks', clothes: 'Clothes', gifts: 'Gifts', games: 'Games', toys: 'Toys', other: 'Other' };
    const septBars = septCats.slice(0, 4).map(c => ({ label: LABEL[c.category], percent: c.percent, tone: TONE[c.category] }));
    const flowNodes = [
      N(60, 300, 'B1 · Home (part 2)', { btn: 1 }),
      N(300, 300, 'C1 · Recap 1, cover', { cardTop: '90px', cardBottom: '30px' }),
      N(540, 300, 'C2 · Kept and streak', { bg: G.yellow, cardTop: '80px', cardBottom: '30px' }),
      N(780, 300, 'C3 · Target and spending', { bg: G.coral, cardTop: '60px', cardBottom: '30px' }),
      N(1020, 300, 'C4 · Archetype', { bg: G.green, cardTop: '110px', cardBottom: '24px', btn: 1 }),
      N(1260, 300, 'C5 · Share sheet', { bg: G.green, cardTop: '110px', cardBottom: '24px', btn: 1, dim: 1, sheetH: '90px' }),
      N(1500, 300, 'C6 · Share image', { bg: G.green, cardTop: '150px', cardBottom: '30px' }),
      N(300, 620, 'C7 · Past recaps', { cardTop: '60px', cardBottom: '90px' }),
      N(540, 620, 'P1 · Month so far, cover', { cardTop: '90px', cardBottom: '30px' }),
      N(780, 620, 'P2 · Streak', { bg: G.yellow, cardTop: '100px', cardBottom: '30px' }),
      N(1020, 620, 'P3 · Spending', { bg: G.coral, cardTop: '60px', cardBottom: '30px' }),
      N(1260, 620, 'P4 · Closing', { bg: G.green, cardTop: '150px', cardBottom: '30px' }),
    ];
    const flowLabels = [
      L(227, 380, 'See last month'), L(230, 585, 'Past recaps'), L(480, 585, 'tap a month'),
      L(467, 404, 'swipe'), L(707, 404, 'swipe'), L(947, 404, 'swipe'),
      L(1187, 404, 'Share your Squirrel'), L(1427, 404, 'the rendered image'), L(588, 200, 'close · from any of the four'),
      L(330, 900, 'preview card · before the first recap'), L(467, 740, 'a "Partial" month'),
      L(707, 724, 'swipe'), L(947, 724, 'swipe'), L(1187, 724, 'swipe'), L(700, 560, 'close'),
      L(1187, 300, 'no Archetype · Share hidden', 'var(--mmr-palette-ink-500)'),
    ];
    return { ...common, septCats, octCats, septBars, flowNodes, flowLabels, septStreak: [12, 20],
      nullValue: null, emptyList: [],
      septRows: [{ label: 'Saved', value: 640, display: '£640', tone: 'green' }, { label: 'Spent', value: 1760, display: '£1,760', tone: 'coral' }],
      keptRows: [{ label: 'Kept', value: 640, display: '£640', tone: 'green' }, { label: 'Target', value: 480, display: '£480', tone: 'violet' }],
      autoScreen: { style: { height: 'auto', minHeight: 0 } },
      n9: 9, n160: 160, n240: 240, n260: 260, n300: 300, n310: 310, n640: 640, n1760: 1760, n2090: 2090, n2400: 2400 };
  }`,
  body: page('My Money Recap · Recap and share (revised)',
    'Part 3, redrawn against SCREEN-STATES.md. Every Recap screen gains the D46 close control; screens 2–4 are rebuilt on the shipped components rather than hand-composed; the "no salary" cover is retired in favour of the Partial Recap (P1–P4). New here: the no-target screen 3, the no-Archetype screen 4, three more share images, the empty past-recaps list and the D66 accessibility-size demonstration. Ledger: Maya, September (target met) and October (target missed), as in the original canvas. Yellow stickies are annotations, not UI.',
    [row('Row 1 · September, target met', [c1, c2, c3, c4, c7empty]),
     row('Row 2 · States and variants', [c2NoTarget, c3Missed, c3NoSpends, c3NoTarget, c1Retired]),
     row('Row 3 · Screen 4 by Archetype, and without one', [c4Treat, c4Hustler, c4Slow, c4NoArchetype]),
     row('Row 4 · Share images for the other three Archetypes', [c6Hustler, c6Treat, c6Slow]),
     row('Row 5 · Accessibility text size and the flow', [a11y, flowC])]),
};
