import { Screen } from '../components/Screen';
import { Chip, ChipRow } from '../components/Chip';
import { Headline, SubCopy, HelperText, Eyebrow } from '../components/Headline';
import { Card, CardTitle } from '../components/Card';
import { Illustration } from '../components/Illustration';
import { StatHero } from '../components/StatHero';
import { StatChip, StatChipRow } from '../components/StatChip';
import { StreakBadge, StreakDots } from '../components/Streak';
import { ProgressBar, CategoryBar, CompareBars } from '../components/Bars';
import { Button } from '../components/Button';
import { ArchetypePill, ArchetypeRow } from '../components/ArchetypePill';
import { PageDots } from '../components/PageDots';
import { Icon } from '../icons';
import { ILLUSTRATIONS } from '../assets/illustrations';
import { CATEGORY_LABEL, CATEGORY_TONE, type Archetype, type Category } from '../components/types';

/* The four monthly-recap screens, composed from the system. Props carry the data; layout is fixed.
   These are the reference compositions for building the rest of the app flow.
   Revised per PRD §13: the cover drops the Earned chip (D2), screen 2 is Kept + streak rather than
   chores (D23), screen 3 uses the piggy marker and all eight Categories (D24), and screen 4 switches
   its mascot per Archetype. */

/** One-line Archetype descriptions, verbatim from PRD §8.6. */
export const ARCHETYPE_BLURB: Record<Archetype, string> = {
  squirrel: "You stashed more than you meant to, and you didn't crack once.",
  hustler: 'You hit your number while still living a little.',
  treatYourself: 'One big treat tipped the month. Worth it?',
  slowBurner: 'Lots of little spends added up. Sneaky.',
};

export interface RecapCoverScreenProps {
  month?: string; childName?: string; total?: number; saved?: number; spent?: number; animate?: boolean;
}
/** Screen 1. Money In is the one big number; the Earned chip is gone because Salary is the only income (D2). */
export function RecapCoverScreen({ month = 'September 2026', childName = 'Maya', total = 2400, saved = 640, spent = 1760, animate = false }: RecapCoverScreenProps) {
  return (
    <Screen ground="recap">
      <ChipRow><Chip asLabel>{month}</Chip><Chip asLabel variant="highlight">{childName}'s recap</Chip></ChipRow>
      <Headline size="l" style={{ marginTop: 22 }}>Your {month.split(' ')[0]} in money</Headline>
      <Illustration src={ILLUSTRATIONS.cover} width={310} />
      <Card underIllustration animateIn={animate}>
        <StatHero eyebrow="Money that came your way" value={total} animate={animate} format={(n) => `£${Math.round(n).toLocaleString('en-GB')}`} />
        <StatChipRow>
          <StatChip tone="saved">Saved £{saved.toLocaleString('en-GB')}</StatChip>
          <StatChip tone="spent">Spent £{spent.toLocaleString('en-GB')}</StatChip>
        </StatChipRow>
        <HelperText>Swipe to see where it all went.</HelperText>
      </Card>
      <div style={{ marginTop: 18 }}><PageDots count={4} active={0} /></div>
    </Screen>
  );
}

export interface RecapEarningScreenProps {
  month?: string;
  /** Money In for the month. */
  moneyIn?: number;
  /** Saved = Money In minus Spent (D5). */
  saved?: number;
  /** Saving Target amount in currency. Pass `null` for a Member who has not set one (§6.12).
   *  It must be `null`, not `undefined` — `undefined` is replaced by the default below. */
  targetAmount?: number | null;
  streakDays?: number;
  streakRange?: [number, number];
  streakCaption?: string;
  daysInMonth?: number;
  animate?: boolean;
}
/** Screen 2. What was kept against the Saving Target, plus the longest no-spend Streak (D23). */
export function RecapEarningScreen({
  month = 'September', moneyIn = 2400, saved = 640, targetAmount = 480,
  streakDays = 9, streakRange = [12, 20], streakCaption = '12 to 20 September. New badge unlocked.',
  daysInMonth = 30, animate = false,
}: RecapEarningScreenProps) {
  const noTarget = targetAmount == null || targetAmount <= 0;
  const keptPct = noTarget ? undefined : Math.min(999, Math.round((saved / targetAmount) * 100));
  const money = (n: number) => `£${n.toLocaleString('en-GB')}`;
  return (
    <Screen ground="earning">
      <ChipRow><Chip asLabel variant="inverse">Kept</Chip><Chip asLabel>{month}</Chip></ChipRow>
      <Headline size="compact" style={{ marginTop: 10 }}>
        {keptPct === undefined
          ? `You kept ${money(saved)} of ${money(moneyIn)}.`
          : `You kept ${money(saved)} of ${money(moneyIn)}. That's ${keptPct}% of your target.`}
      </Headline>
      <Illustration src={ILLUSTRATIONS.jar} width={190} style={{ marginBottom: -10 }} />
      <Card style={{ gap: 12 }} animateIn={animate}>
        {noTarget ? (
          <>
            <Eyebrow>Saving target</Eyebrow>
            <CardTitle>Set a target to see how you did</CardTitle>
            <HelperText>Pick a share of your salary to keep each month, and next month's recap will score it.</HelperText>
          </>
        ) : (
          <>
            <Eyebrow>Kept against your target</Eyebrow>
            <CompareBars
              animate={animate}
              scale="max"
              rows={[
                { label: 'Kept', value: saved, display: money(saved), tone: 'green' },
                { label: 'Target', value: targetAmount, display: money(targetAmount), tone: 'violet' },
              ]}
            />
          </>
        )}
      </Card>
      <Card variant="accent" style={{ marginTop: 12 }} animateIn={animate}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 14 }}>
          <StreakBadge value={streakDays} />
          <div style={{ display: 'grid', gap: 2 }}>
            <CardTitle>Longest no-spend streak</CardTitle>
            <div style={{ fontSize: 13, fontWeight: 700, opacity: 0.8 }}>{streakCaption}</div>
          </div>
        </div>
        <StreakDots total={daysInMonth} activeRange={streakRange} />
      </Card>
      <div style={{ marginTop: 18 }}><PageDots count={4} active={1} onLight /></div>
    </Screen>
  );
}

export interface RecapGoalScreenProps {
  /** Saved for the month. */
  saved?: number;
  /** Saving Target amount. */
  target?: number;
  /** One sentence from PRD §8.7. */
  helper?: string;
  subCopy?: string;
  spentTotal?: number;
  /** Every Category with a Spend, sorted descending by percent of Spent. */
  categories?: Array<{ category: Category; percent: number }>;
  /** Bars drawn before the tail is rolled into one line. Four is what the split card holds at 844pt. */
  maxCategoryBars?: number;
  /** The largest single Spend. Pass `null` when the month has no Spends (§6.13); like `targetAmount`,
   *  `undefined` would be replaced by the default below. It is hidden automatically when Spent is 0. */
  splurge?: string | null;
  animate?: boolean;
}
/** Screen 3. Saving-target progress with the piggy marker (D24), then where the money went, by Category. */
export function RecapGoalScreen({
  saved = 640, target = 480, helper = 'Target met. £160 to spare.', subCopy = 'You kept a third more than you set out to keep.',
  spentTotal = 1760, animate = false,
  categories = [
    { category: 'bills', percent: 35 },
    { category: 'rides', percent: 26 },
    { category: 'snacks', percent: 14 },
    { category: 'clothes', percent: 10 },
    { category: 'gifts', percent: 7 },
    { category: 'games', percent: 5 },
    { category: 'toys', percent: 2 },
    { category: 'other', percent: 1 },
  ],
  splurge = '£420 on a flight home. Your biggest single buy.',
  maxCategoryBars = 4,
}: RecapGoalScreenProps) {
  const pct = Math.min(999, Math.round((saved / target) * 100));
  const money = (n: number) => `£${n.toLocaleString('en-GB')}`;
  const noSpends = spentTotal <= 0 || categories.length === 0;
  const shown = categories.slice(0, maxCategoryBars);
  const rest = categories.slice(maxCategoryBars);
  const restPct = rest.reduce((sum, c) => sum + c.percent, 0);
  return (
    <Screen ground="goal">
      <ChipRow><Chip asLabel>Target</Chip><Chip asLabel variant="inverse">Spending</Chip></ChipRow>
      <Headline size="compact" style={{ marginTop: 16 }}>{pct}% of your saving target</Headline>
      <SubCopy style={{ marginTop: 8 }}>{subCopy}</SubCopy>
      <Card style={{ marginTop: 18 }} animateIn={animate}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline' }}>
          <CardTitle>Saving target</CardTitle>
          <span style={{ fontSize: 14, fontWeight: 800, color: 'var(--mmr-color-text-secondary)' }}>
            <span style={{ color: 'var(--mmr-color-text-primary)', fontSize: 18 }}>{money(saved)}</span> of {money(target)}
          </span>
        </div>
        <div style={{ paddingTop: 14 }}><ProgressBar value={saved} max={target} animate={animate} marker={<Icon name="piggy" />} /></div>
        <HelperText>{helper}</HelperText>
      </Card>
      <Card variant="split" style={{ marginTop: 12, flexGrow: 1 }} animateIn={animate}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 10, flexGrow: 1, minWidth: 0 }}>
          <div>
            <Eyebrow>{noSpends ? 'Where it went' : `Where ${money(spentTotal)} went`}</Eyebrow>
            <CardTitle>{noSpends ? 'Nothing spent this month' : `Mostly ${CATEGORY_LABEL[categories[0]!.category].toLowerCase()}`}</CardTitle>
          </div>
          <div style={{ display: 'grid', gap: 6 }}>
            {shown.map((c) => (
              <CategoryBar key={c.category} label={CATEGORY_LABEL[c.category]} percent={c.percent} tone={CATEGORY_TONE[c.category]} animate={animate} />
            ))}
            {noSpends ? <HelperText>You kept every penny that came in.</HelperText> : null}
            {rest.length ? (
              <HelperText>
                {rest.map((c) => CATEGORY_LABEL[c.category]).join(', ')} · {restPct}%
              </HelperText>
            ) : null}
          </div>
          {splurge && !noSpends ? (
            <div style={{ marginTop: 'auto', paddingTop: 4, display: 'grid', gap: 2 }}>
              <Eyebrow tone="coral">The splurge</Eyebrow>
              <div style={{ fontSize: 15, fontWeight: 800, lineHeight: 1.3 }}>{splurge}</div>
            </div>
          ) : null}
        </div>
        <div style={{ width: 148, flexShrink: 0, display: 'flex', alignItems: 'flex-end', justifyContent: 'center', marginTop: -60 }}>
          <img src={ILLUSTRATIONS.vending} alt="" style={{ width: 148, display: 'block' }} />
        </div>
      </Card>
      <div style={{ marginTop: 18 }}><PageDots count={4} active={2} /></div>
    </Screen>
  );
}

/** The mascot poster for each Archetype. Prompts: my-money-recap/illustration-prompts.md. */
const ARCHETYPE_ILLUSTRATION: Record<Archetype, string> = {
  squirrel: ILLUSTRATIONS.squirrel,
  hustler: ILLUSTRATIONS.hustler,
  treatYourself: ILLUSTRATIONS.treat,
  slowBurner: ILLUSTRATIONS.slowburner,
};
const ARCHETYPE_TITLE: Record<Archetype, string> = {
  squirrel: 'The Squirrel', hustler: 'The Hustler', treatYourself: 'Treat-Yourself', slowBurner: 'The Slow Burner',
};

export interface RecapPersonalityScreenProps {
  childName?: string; archetype?: Archetype; archetypeLabel?: string; blurb?: string; saved?: number; spent?: number; animate?: boolean;
}
/** Screen 4. The shareable Archetype badge. No page dots, so screenshots stay clean. */
export function RecapPersonalityScreen({
  childName = 'Maya', archetype = 'squirrel', archetypeLabel, blurb, saved = 640, spent = 1760, animate = false,
}: RecapPersonalityScreenProps) {
  const all: Archetype[] = ['squirrel', 'hustler', 'treatYourself', 'slowBurner'];
  const title = archetypeLabel ?? ARCHETYPE_TITLE[archetype];
  const money = (n: number) => `£${n.toLocaleString('en-GB')}`;
  return (
    <Screen ground="personality">
      <ChipRow><Chip asLabel>Your money personality</Chip></ChipRow>
      <div style={{ marginTop: 16, fontSize: 15, fontWeight: 800, opacity: 0.85 }}>This month, {childName}, you were</div>
      <Headline size="display" style={{ marginTop: 2 }}>{title}</Headline>
      <SubCopy style={{ marginTop: 12, opacity: 0.9 }}>{blurb ?? ARCHETYPE_BLURB[archetype]}</SubCopy>
      <Illustration src={ARCHETYPE_ILLUSTRATION[archetype]} width={340} maxHeight={300} />
      <Card underIllustration style={{ gap: 14 }} animateIn={animate}>
        <CompareBars animate={animate} rows={[{ label: 'Saved', value: saved, display: money(saved), tone: 'green' }, { label: 'Spent', value: spent, display: money(spent), tone: 'coral' }]} />
        <Button icon={<Icon name="share" />}>Share your {title.replace(/^The /, '')}</Button>
      </Card>
      <div style={{ marginTop: 12, marginLeft: 18 }}>
        <ArchetypeRow>{all.map((a) => <ArchetypePill key={a} archetype={a} selected={a === archetype} />)}</ArchetypeRow>
      </div>
    </Screen>
  );
}
