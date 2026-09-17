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
import { PromptCard } from '../components/PromptCard';
import { moneyFormat } from '../money';
import { PageDots } from '../components/PageDots';
import { Icon } from '../icons';
import { ILLUSTRATIONS } from '../assets/illustrations';
import { CATEGORY_LABEL, CATEGORY_TONE, type Archetype, type Category } from '../components/types';

/* The four monthly-recap screens, composed from the system and revised per PRD §13 (D2, D23, D24).
   Props carry the data; layout is fixed. These are the reference compositions for the rest of the app. */

/** One-line Archetype descriptions, verbatim from PRD §8.6. */
export const ARCHETYPE_BLURB: Record<Archetype, string> = {
  squirrel: "You stashed more than you meant to, and you didn't crack once.",
  hustler: 'You hit your number while still living a little.',
  treatYourself: 'One big treat tipped the month. Worth it?',
  slowBurner: 'Lots of little spends added up. Sneaky.',
};

/** Every Recap screen formats money the same way (PRD D45, D71): `currency` is the Member's ISO 4217
 *  code, `locale` the viewer's language tag (omit for the device language). See `../money.ts`. The
 *  sample defaults are Maya's September (PRD §8), in pounds. */
export interface MoneyProps {
  currency?: string;
  locale?: string;
}
const useMoney = ({ currency = 'GBP', locale }: MoneyProps) => moneyFormat(currency, locale);

export interface RecapCoverScreenProps extends MoneyProps {
  month?: string; childName?: string; total?: number; saved?: number; spent?: number; animate?: boolean;
}
/** Screen 1. Money In is the one big number; the Earned chip is gone because Salary is the only income (D2). */
export function RecapCoverScreen({ month = 'September 2026', childName = 'Maya', total = 2400, saved = 640, spent = 1760, animate = false, currency, locale }: RecapCoverScreenProps) {
  const format = useMoney({ currency, locale });
  return (
    <Screen ground="recap">
      <ChipRow><Chip asLabel>{month}</Chip><Chip asLabel variant="highlight">{childName}'s recap</Chip></ChipRow>
      <Headline size="l" style={{ marginTop: 22 }}>Your {month.split(' ')[0]} in money</Headline>
      <Illustration src={ILLUSTRATIONS.cover} width={310} />
      <Card underIllustration animateIn={animate}>
        <StatHero eyebrow="Money that came your way" value={total} animate={animate} format={format} />
        <StatChipRow>
          <StatChip tone="saved">Saved {format(saved)}</StatChip>
          <StatChip tone="spent">Spent {format(spent)}</StatChip>
        </StatChipRow>
        <HelperText>Swipe to see where it all went.</HelperText>
      </Card>
      <div style={{ marginTop: 18 }}><PageDots count={4} active={0} /></div>
    </Screen>
  );
}

export interface RecapEarningScreenProps extends MoneyProps {
  month?: string;
  /** Money In for the month. */
  moneyIn?: number;
  /** Saved = Money In minus Spent (D5). */
  saved?: number;
  /** Saving Target amount in currency. Pass `null` for a Member who has not set one (§6.12).
   *  It must be `null`, not `undefined` — `undefined` is replaced by the default below. */
  targetAmount?: number | null;
  /** Opens Settings › Saving target from the no-target prompt. */
  onSetTarget?: () => void;
  streakDays?: number;
  streakRange?: [number, number];
  streakCaption?: string;
  daysInMonth?: number;
  animate?: boolean;
}
/** Screen 2. What was kept against the Saving Target, plus the longest no-spend Streak (D23). */
export function RecapEarningScreen({
  month = 'September', moneyIn = 2400, saved = 640, targetAmount = 480,
  onSetTarget, streakDays = 9, streakRange = [12, 20], streakCaption = '12 to 20 September. New badge unlocked.',
  daysInMonth = 30, animate = false, currency, locale,
}: RecapEarningScreenProps) {
  const money = useMoney({ currency, locale });
  const noTarget = targetAmount == null || targetAmount <= 0;
  const keptPct = noTarget ? undefined : Math.min(999, Math.round((saved / targetAmount) * 100));
  return (
    <Screen ground="earning">
      <ChipRow><Chip asLabel variant="inverse">Kept</Chip><Chip asLabel>{month}</Chip></ChipRow>
      <Headline size="compact" style={{ marginTop: 10 }}>
        {keptPct === undefined
          ? `You kept ${money(saved)} of ${money(moneyIn)}.`
          : `You kept ${money(saved)} of ${money(moneyIn)}. That's ${keptPct}% of your target.`}
      </Headline>
      <Illustration src={ILLUSTRATIONS.jar} width={190} style={{ marginBottom: -10 }} />
      {noTarget ? (
        <PromptCard eyebrow="Saving target" title="Set a target to see how you did" helper="Pick a share of your salary to keep each month, and next month's recap will score it." onClick={onSetTarget} />
      ) : (
        <Card style={{ gap: 12 }} animateIn={animate}>
          <Eyebrow>Kept against your target</Eyebrow>
          <CompareBars
            animate={animate}
            scale="max"
            rows={[
              { label: 'Kept', value: saved, display: money(saved), tone: 'green' },
              { label: 'Target', value: targetAmount, display: money(targetAmount), tone: 'violet' },
            ]}
          />
        </Card>
      )}
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

export interface RecapGoalScreenProps extends MoneyProps {
  /** Saved for the month. */
  saved?: number;
  /** Saving Target amount, or `null` for a Member who has not set one (SCREEN-STATES C3): the headline
   *  becomes "Where {Spent} went" and the target card becomes a `PromptCard`. `null`, not `undefined`
   *  — `undefined` takes the default. */
  target?: number | null;
  /** One sentence from PRD §8.7. Unused when `target` is null. Defaults to the sample month's sentence in `currency`. */
  helper?: string;
  /** Opens Settings › Saving target from the no-target prompt. */
  onSetTarget?: () => void;
  subCopy?: string;
  spentTotal?: number;
  /** Every Category with a Spend, sorted descending by percent of Spent. */
  categories?: Array<{ category: Category; percent: number }>;
  /** Bars drawn before the tail is rolled into one line. Three is what the split card holds at 844pt in
   *  every Currency (PRD §6.13). */
  maxCategoryBars?: number;
  /** The largest single Spend. Pass `null` when the month has no Spends (§6.13); like `targetAmount`,
   *  `undefined` would be replaced by the default below. It is hidden automatically when Spent is 0. */
  splurge?: string | null;
  animate?: boolean;
}
/** Screen 3. Saving-target progress with the piggy marker (D24), then where the money went, by Category. */
export function RecapGoalScreen({
  saved = 640, target = 480, helper, subCopy = 'You kept a third more than you set out to keep.',
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
  splurge, maxCategoryBars = 3, onSetTarget, currency, locale,
}: RecapGoalScreenProps) {
  const money = useMoney({ currency, locale });
  const noTarget = target == null || target <= 0;
  const pct = noTarget ? 0 : Math.min(999, Math.round((saved / target) * 100));
  // Sample defaults built from the formatter, so a rupiah preview never shows a pound sign.
  const helperText = helper ?? (noTarget ? '' : `Target met. ${money(saved - target)} to spare.`);
  const splurgeText = splurge === undefined ? `${money(420)} on a flight home. Your biggest single buy.` : splurge;
  const bars = maxCategoryBars;
  const noSpends = spentTotal <= 0 || categories.length === 0;
  const shown = categories.slice(0, bars);
  const rest = categories.slice(bars);
  const restPct = rest.reduce((sum, c) => sum + c.percent, 0);
  return (
    <Screen ground="goal">
      <ChipRow><Chip asLabel>Target</Chip><Chip asLabel variant="inverse">Spending</Chip></ChipRow>
      {noTarget ? (
        <Headline size="compact" style={{ marginTop: 16 }}>Where {money(spentTotal)} went</Headline>
      ) : (
        <>
          <Headline size="compact" style={{ marginTop: 16 }}>{pct}% of your saving target</Headline>
          <SubCopy style={{ marginTop: 8 }}>{subCopy}</SubCopy>
        </>
      )}
      {noTarget ? (
        <div style={{ marginTop: 18 }}>
          <PromptCard title="Set a saving target" helper="Pick a share of your salary to keep each month." onClick={onSetTarget} />
        </div>
      ) : (
        <Card style={{ marginTop: 18 }} animateIn={animate}>
          <div style={{ display: 'flex', flexWrap: 'wrap', justifyContent: 'space-between', alignItems: 'baseline', columnGap: 12, rowGap: 4 }}>
            <CardTitle>Saving target</CardTitle>
            <span style={{ fontSize: 14, fontWeight: 800, color: 'var(--mmr-color-text-secondary)' }}>
              <span style={{ color: 'var(--mmr-color-text-primary)', fontSize: 18 }}>{money(saved)}</span> of {money(target)}
            </span>
          </div>
          <div style={{ paddingTop: 14 }}><ProgressBar value={saved} max={target} animate={animate} marker={<Icon name="piggy" />} /></div>
          <HelperText>{helperText}</HelperText>
        </Card>
      )}
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
          {splurgeText && !noSpends ? (
            <div style={{ marginTop: 'auto', paddingTop: 4, display: 'grid', gap: 2 }}>
              <Eyebrow tone="coral">The splurge</Eyebrow>
              <div style={{ fontSize: 15, fontWeight: 800, lineHeight: 1.3 }}>{splurgeText}</div>
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

export interface RecapPersonalityScreenProps extends MoneyProps {
  childName?: string; archetype?: Archetype; archetypeLabel?: string; blurb?: string; saved?: number; spent?: number; animate?: boolean;
}
/** Screen 4. The shareable Archetype badge. No page dots, so screenshots stay clean. */
export function RecapPersonalityScreen({
  childName = 'Maya', archetype = 'squirrel', archetypeLabel, blurb, saved = 640, spent = 1760, animate = false, currency, locale,
}: RecapPersonalityScreenProps) {
  const money = useMoney({ currency, locale });
  const all: Archetype[] = ['squirrel', 'hustler', 'treatYourself', 'slowBurner'];
  const title = archetypeLabel ?? ARCHETYPE_TITLE[archetype];
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
