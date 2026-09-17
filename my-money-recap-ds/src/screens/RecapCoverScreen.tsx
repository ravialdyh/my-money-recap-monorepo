import { ILLUSTRATIONS } from '../assets/illustrations';
import { Card } from '../components/Card';
import { Chip, ChipRow } from '../components/Chip';
import { Headline, HelperText } from '../components/Headline';
import { Illustration } from '../components/Illustration';
import { PageDots } from '../components/PageDots';
import { Screen } from '../components/Screen';
import { StatChip, StatChipRow } from '../components/StatChip';
import { StatHero } from '../components/StatHero';
import { type MoneyProps, useMoney } from './shared';

/* One of the four monthly-recap screens, composed from the system and revised per PRD §13.
   Props carry the data; layout is fixed. These are the reference compositions for the rest of the app. */

export interface RecapCoverScreenProps extends MoneyProps {
  month?: string;
  childName?: string;
  total?: number;
  saved?: number;
  spent?: number;
  animate?: boolean;
}
/** Screen 1. Money In is the one big number; the Earned chip is gone because Salary is the only income (D2). */
export function RecapCoverScreen({
  month = 'September 2026',
  childName = 'Maya',
  total = 2400,
  saved = 640,
  spent = 1760,
  animate = false,
  currency,
  locale,
}: RecapCoverScreenProps) {
  const format = useMoney({ currency, locale });
  return (
    <Screen ground="recap">
      <ChipRow>
        <Chip asLabel>{month}</Chip>
        <Chip asLabel variant="highlight">
          {childName}'s recap
        </Chip>
      </ChipRow>
      <Headline size="l" style={{ marginTop: 22 }}>
        Your {month.split(' ')[0]} in money
      </Headline>
      <Illustration src={ILLUSTRATIONS.cover} width={310} />
      <Card underIllustration animateIn={animate}>
        <StatHero eyebrow="Money that came your way" value={total} animate={animate} format={format} />
        <StatChipRow>
          <StatChip tone="saved">Saved {format(saved)}</StatChip>
          <StatChip tone="spent">Spent {format(spent)}</StatChip>
        </StatChipRow>
        <HelperText>Swipe to see where it all went.</HelperText>
      </Card>
      <div style={{ marginTop: 18 }}>
        <PageDots count={4} active={0} />
      </div>
    </Screen>
  );
}
