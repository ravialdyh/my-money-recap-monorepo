import { ILLUSTRATIONS } from '../assets/illustrations';
import { CategoryBar, ProgressBar } from '../components/Bars';
import { Card, CardTitle } from '../components/Card';
import { Chip, ChipRow } from '../components/Chip';
import { Eyebrow, Headline, HelperText, SubCopy } from '../components/Headline';
import { PageDots } from '../components/PageDots';
import { PromptCard } from '../components/PromptCard';
import { Screen } from '../components/Screen';
import { CATEGORY_LABEL, CATEGORY_TONE, type Category } from '../components/types';
import { Icon } from '../icons';
import { type MoneyProps, useMoney } from './shared';

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
  saved = 640,
  target = 480,
  helper,
  subCopy = 'You kept a third more than you set out to keep.',
  spentTotal = 1760,
  animate = false,
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
  splurge,
  maxCategoryBars = 3,
  onSetTarget,
  currency,
  locale,
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
      <ChipRow>
        <Chip asLabel>Target</Chip>
        <Chip asLabel variant="inverse">
          Spending
        </Chip>
      </ChipRow>
      {noTarget ? (
        <Headline size="compact" style={{ marginTop: 16 }}>
          Where {money(spentTotal)} went
        </Headline>
      ) : (
        <>
          <Headline size="compact" style={{ marginTop: 16 }}>
            {pct}% of your saving target
          </Headline>
          <SubCopy style={{ marginTop: 8 }}>{subCopy}</SubCopy>
        </>
      )}
      {noTarget ? (
        <div style={{ marginTop: 18 }}>
          <PromptCard
            title="Set a saving target"
            helper="Pick a share of your salary to keep each month."
            onClick={onSetTarget}
          />
        </div>
      ) : (
        <Card style={{ marginTop: 18 }} animateIn={animate}>
          <div
            style={{
              display: 'flex',
              flexWrap: 'wrap',
              justifyContent: 'space-between',
              alignItems: 'baseline',
              columnGap: 12,
              rowGap: 4,
            }}
          >
            <CardTitle>Saving target</CardTitle>
            <span style={{ fontSize: 14, fontWeight: 800, color: 'var(--mmr-color-text-secondary)' }}>
              <span style={{ color: 'var(--mmr-color-text-primary)', fontSize: 18 }}>{money(saved)}</span> of{' '}
              {money(target)}
            </span>
          </div>
          <div style={{ paddingTop: 14 }}>
            <ProgressBar value={saved} max={target} animate={animate} marker={<Icon name="piggy" />} />
          </div>
          <HelperText>{helperText}</HelperText>
        </Card>
      )}
      <Card variant="split" style={{ marginTop: 12, flexGrow: 1 }} animateIn={animate}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 10, flexGrow: 1, minWidth: 0 }}>
          <div>
            <Eyebrow>{noSpends ? 'Where it went' : `Where ${money(spentTotal)} went`}</Eyebrow>
            <CardTitle>
              {noSpends
                ? 'Nothing spent this month'
                : `Mostly ${CATEGORY_LABEL[categories[0]!.category].toLowerCase()}`}
            </CardTitle>
          </div>
          <div style={{ display: 'grid', gap: 6 }}>
            {shown.map((c) => (
              <CategoryBar
                key={c.category}
                label={CATEGORY_LABEL[c.category]}
                percent={c.percent}
                tone={CATEGORY_TONE[c.category]}
                animate={animate}
              />
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
        <div
          style={{
            width: 148,
            flexShrink: 0,
            display: 'flex',
            alignItems: 'flex-end',
            justifyContent: 'center',
            marginTop: -60,
          }}
        >
          <img src={ILLUSTRATIONS.vending} alt="" style={{ width: 148, display: 'block' }} />
        </div>
      </Card>
      <div style={{ marginTop: 18 }}>
        <PageDots count={4} active={2} />
      </div>
    </Screen>
  );
}
