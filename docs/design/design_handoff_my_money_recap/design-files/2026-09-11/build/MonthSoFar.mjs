import { X, col, row, page, screen, withClose, logicCommon } from './lib.mjs';

// P1 · cover. Spent is the hero; the only chip is the spend count. No Money In anywhere.
const p1 = (label, headline, spentKey, spentText, chipText, notesArr) => col(label, withClose(screen('recap', `${X('ChipRow', '', `${X('Chip', 'as-label="{{ true }}"', 'September 2026', 'auto,30px')}${X('Chip', 'as-label="{{ true }}" variant="highlight"', "Maya's recap", 'auto,30px')}`, '100%,30px')}
  <div style="margin-top:22px">${X('Headline', 'size="l"', headline, '100%,98px')}</div>
  ${X('Illustration', 'src="{{ coverSrc }}" width="{{ n310 }}"', '', '100%,260px')}
  ${X('Card', 'under-illustration="{{ true }}"', `${X('StatHero', `eyebrow="What you've spent" value="{{ ${spentKey} }}" format="{{ gbp }}"`, '', '100%,100px')}
    ${X('StatChipRow', '', X('StatChip', 'tone="spent"', chipText, 'auto,34px'), '100%,34px')}
    ${X('HelperText', '', 'Swipe to see where it all went.', '100%,20px')}`, '100%,230px')}
  <div style="margin-top:18px">${X('PageDots', 'count="{{ n4 }}" active="{{ n0 }}" on-select="{{ noop }}"', '', 'auto,6px')}</div>`)), notesArr);

// P2 · streak. StreakBadge, StreakDots, the streak sentence. No CompareBars.
const p2 = col('P2 · Streak', withClose(screen('earning', `${X('ChipRow', '', `${X('Chip', 'as-label="{{ true }}" variant="inverse"', 'Streak', 'auto,30px')}${X('Chip', 'as-label="{{ true }}"', 'September', 'auto,30px')}`, '100%,30px')}
  <div style="margin-top:10px">${X('Headline', 'size="compact"', 'Your longest no-spend streak', '100%,80px')}</div>
  ${X('Illustration', 'src="{{ jarSrc }}" width="{{ n190 }}" dc-props="{{ jarStyle }}"', '', '100%,190px')}
  ${X('Card', 'variant="accent"', `<div style="display:flex;align-items:center;gap:14px">
      ${X('StreakBadge', 'value="{{ n9 }}"', '', '64px,64px')}
      <div style="display:grid;gap:2px">
        ${X('CardTitle', '', '12 to 20 September', 'auto,22px')}
        <div style="font-size:13px;font-weight:700;opacity:.8">New badge unlocked.</div>
      </div>
    </div>
    ${X('StreakDots', 'total="{{ n20 }}" active-range="{{ streakRange }}"', '', '100%,36px')}`, '100%,170px')}
  <div style="margin-top:18px">${X('PageDots', 'count="{{ n4 }}" active="{{ n1 }}" on-light="{{ true }}" on-select="{{ noop }}"', '', 'auto,6px')}</div>`), true), [
  'Same layout as C2 minus the compare card: with Money In 0 there is nothing to hold Kept against, so the streak is the whole screen. Streak = longest run of spend-free days inside the window (PRD §8.5). Maya signed up before September, so the window is 1–20 September on the 20th; nothing is dated 12–20 → 9 days, ≥ 7 → "New badge unlocked." The dot strip stops at today (20 dots, not 30): the app shows no day it has not observed.',
  'The headline is the card title lifted up, so the card carries only the badge, the date range and the badge copy, the two strings SCREEN-STATES fixes. Yellow ground: the close control is ink, PageDots take onLight.',
]);

// P3 · spending. Before payday the target card is replaced by "Your salary lands on the 25th." (D55).
const p3 = col('P3 · Spending', withClose(screen('goal', `${X('ChipRow', '', `${X('Chip', 'as-label="{{ true }}"', 'Target', 'auto,30px')}${X('Chip', 'as-label="{{ true }}" variant="inverse"', 'Spending', 'auto,30px')}`, '100%,30px')}
  <div style="margin-top:16px">${X('Headline', 'size="compact"', 'Where £900 went', '100%,42px')}</div>
  <div style="margin-top:18px">${X('Card', '', `${X('Eyebrow', '', 'Saving target', 'auto,16px')}${X('CardTitle', '', 'Your salary lands on the 25th.', '100%,22px')}`, '100%,90px')}</div>
  <div style="margin-top:12px;flex-grow:1;display:flex">
  ${X('Card', 'variant="split" dc-props="{{ growCard }}"', `<div style="display:flex;flex-direction:column;gap:10px;flex-grow:1;min-width:0">
      <div>${X('Eyebrow', '', 'Where £900 went', 'auto,16px')}${X('CardTitle', '', 'Mostly bills', '100%,22px')}</div>
      <div style="display:grid;gap:6px">
        <sc-for list="{{ p3Bars }}" as="b" hint-placeholder-count="3">${X('CategoryBar', 'label="{{ b.label }}" percent="{{ b.percent }}" tone="{{ b.tone }}"', '', '100%,40px')}</sc-for>
        ${X('HelperText', '', 'Rides, Other · 6%', '100%,20px')}
      </div>
      <div style="margin-top:auto;padding-top:4px;display:grid;gap:2px">
        ${X('Eyebrow', 'tone="coral"', 'The splurge', 'auto,16px')}
        <div style="font-size:15px;font-weight:800;line-height:1.3">£280 on phone and internet. Your biggest single buy.</div>
      </div>
    </div>
    <div style="width:148px;flex-shrink:0;display:flex;align-items:flex-end;justify-content:center;margin-top:-60px"><img src="{{ vendingSrc }}" alt="" style="width:148px;display:block"></div>`, '100%,420px')}
  </div>
  <div style="margin-top:18px">${X('PageDots', 'count="{{ n4 }}" active="{{ n2 }}" on-select="{{ noop }}"', '', 'auto,6px')}</div>`)), [
  'Same three-bars-plus-roll-up rule and the same Splurge card as C3 (D71). 20 September, £900 across 8 spends: Bills £500 (56%), Clothes £180 (20%), Snacks £160 (18%), Rides £40 and Other £20 rolled into one line (6%). Splurge = largest single spend, £280 phone and internet.',
  'No percentage headline and no SubCopy: there is no target amount to score against. D55: before Payday the saving-target card is the one line "Your salary lands on the 25th." In a past partial month (situation 3) the card is omitted and the split card takes its height.',
]);

// P4 · closing. No mascot, no Archetype, no Share. Pills at 40%. Line by situation.
const p4 = (label, line, notesArr) => col(label, withClose(screen('personality', `${X('ChipRow', '', X('Chip', 'as-label="{{ true }}"', 'Your money personality', 'auto,30px'), '100%,30px')}
  <div style="margin-top:16px">${X('Headline', 'size="m"', line, '100%,84px')}</div>
  <div style="flex-grow:1"></div>
  <div style="display:flex;flex-direction:column;gap:10px;margin-bottom:8px">
    <div style="font:800 12px/1.2 var(--mmr-font-family-text);letter-spacing:.08em;text-transform:uppercase;opacity:.85">Archetypes</div>
    <div style="opacity:.4">${X('ArchetypeRow', '', `${X('ArchetypePill', 'archetype="squirrel"', '', 'auto,24px')}${X('ArchetypePill', 'archetype="hustler"', '', 'auto,24px')}${X('ArchetypePill', 'archetype="treatYourself"', '', 'auto,24px')}${X('ArchetypePill', 'archetype="slowBurner"', '', 'auto,24px')}`, '100%,24px')}</div>
  </div>
  <div style="margin-top:18px">${X('PageDots', 'count="{{ n4 }}" active="{{ n3 }}" on-select="{{ noop }}"', '', 'auto,6px')}</div>`)), notesArr);

const situation1 = [
  p1('P1 · Cover (preview, current month)', 'September so far', 'n900', '£900', '8 spends', [
    'Opened from the Home preview card on 20 September (D47). StatHero "What you\'ve spent" = Spent £900; the one chip counts spends. No Money In figure anywhere: showing the full Salary would credit money the app has not observed yet (D52).',
    'Situation 1 of 3: the current month, still running. Only this headline and P4\'s closing line change between the three situations (D64); P2 and P3 are identical in all three.',
    'D46: 44pt close control top-left, `plus` rotated 45°; the chip row shifts 34pt right to clear it. PageDots are tappable.',
  ]),
  p2, p3,
  p4('P4 · Closing (preview)', 'Your recap arrives on 1 October.', [
    'No mascot, no Archetype, no Share button. The four ArchetypePills sit at 40% so the Member can see what is coming. Read-only: the close control and the dots are the only controls.',
    'Closing line by situation: "Your recap arrives on {1st of next month}." · "Your first full recap arrives on {date}." · "Set your monthly salary to get a full recap." (§6.17).',
  ]),
];

const situation23 = [
  p1('P1 · Cover / sign-up month, payday already passed', 'Your first 4 days', 'n200', '£200', '4 spends', [
    'Situation 2 (§8.11 worked example): sign-up 27 September, Payday 25. September\'s effective payday precedes sign-up, so the month is partial. £200 across four days → "Your first 4 days" over £200. Reached from the Home preview card during the month and from the past-recaps list ("September 2026 · Partial") after it.',
  ]),
  p4('P4 · Closing / sign-up month', 'Your first full recap arrives on 1 November.', [
    'Situation 2 closing line. October is this Member\'s first full month, so its Recap on 1 November is the first with an Archetype.',
  ]),
  p1('P1 · Cover / past month, no salary', 'Your September spending', 'n1760', '£1,760', '15 spends', [
    'Situation 3: any past month with Money In 0 and at least one Spend, here Maya\'s September with no Salary ever set. Spent £1,760 across 15 spends.',
  ]),
  p4('P4 · Closing / past month, no salary', 'Set your monthly salary to get a full recap.', [
    'Situation 3 closing line. The line is a prompt, but this screen stays read-only: the Member sets the Salary in Settings.',
  ]),
];

export default {
  props: { showAnnotations: { editor: 'boolean', default: true, tsType: 'boolean', section: 'Canvas' } },
  logic: `  renderVals() {${logicCommon}
    const p3Bars = [['Bills', 56, 'ink'], ['Clothes', 20, 'green'], ['Snacks', 18, 'coral']].map(([label, percent, tone]) => ({ label, percent, tone }));
    return { ...common, p3Bars, streakRange: [12, 20], n9: 9, n20: 20, n190: 190, n310: 310, n900: 900, n200: 200, n1760: 1760 };
  }`,
  body: page('My Money Recap · Month so far and Partial Recap (P1–P4)',
    'New group, D47 / D52 / D64. Four artboards serve three situations: the month-so-far preview a Member opens before their first Recap, the sign-up month whose Payday came before sign-up, and any past month with Money In 0. The layout is identical in all three; only the screen-1 headline and the screen-4 closing line change, so row 2 draws just those two screens for situations 2 and 3. Read-only: no Archetype, no Share. Ledger: Maya, GBP, Payday 25, Target 20%, seen on 20 September. Yellow stickies are annotations, not UI.',
    [row('Row 1 · The four screens, situation 1 (preview of the running month)', situation1), row('Row 2 · What changes in situations 2 and 3', situation23)]),
};
