import { ILLUSTRATIONS } from '../assets/illustrations';
import { CompareBars } from '../components/Bars';
import { Card, CardTitle } from '../components/Card';
import { Chip, ChipRow } from '../components/Chip';
import { Eyebrow, Headline } from '../components/Headline';
import { Illustration } from '../components/Illustration';
import { PageDots } from '../components/PageDots';
import { PromptCard } from '../components/PromptCard';
import { Screen } from '../components/Screen';
import { StreakBadge, StreakDots } from '../components/Streak';
import { type MoneyProps, useMoney } from './shared';

/* One of the four monthly-recap screens, composed from the system and revised per PRD §13.
   Props carry the data; layout is fixed. These are the reference compositions for the rest of the app. */

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
  month = 'September',
  moneyIn = 2400,
  saved = 640,
  targetAmount = 480,
  onSetTarget,
  streakDays = 9,
  streakRange = [12, 20],
  streakCaption = '12 to 20 September. New badge unlocked.',
  daysInMonth = 30,
  animate = false,
  currency,
  locale,
}: RecapEarningScreenProps) {
  const money = useMoney({ currency, locale });
  const noTarget = targetAmount == null || targetAmount <= 0;
  const keptPct = noTarget ? undefined : Math.min(999, Math.round((saved / targetAmount) * 100));
  return (
    <Screen ground="earning">
      <ChipRow>
        <Chip asLabel variant="inverse">
          Kept
        </Chip>
        <Chip asLabel>{month}</Chip>
      </ChipRow>
      <Headline size="compact" style={{ marginTop: 10 }}>
        {keptPct === undefined
          ? `You kept ${money(saved)} of ${money(moneyIn)}.`
          : `You kept ${money(saved)} of ${money(moneyIn)}. That's ${keptPct}% of your target.`}
      </Headline>
      <Illustration src={ILLUSTRATIONS.jar} width={190} style={{ marginBottom: -10 }} />
      {noTarget ? (
        <PromptCard
          eyebrow="Saving target"
          title="Set a target to see how you did"
          helper="Pick a share of your salary to keep each month, and next month's recap will score it."
          onClick={onSetTarget}
        />
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
      <div style={{ marginTop: 18 }}>
        <PageDots count={4} active={1} onLight />
      </div>
    </Screen>
  );
}
