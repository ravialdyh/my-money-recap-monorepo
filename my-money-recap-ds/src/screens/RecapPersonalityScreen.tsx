import { ILLUSTRATIONS } from '../assets/illustrations';
import { ArchetypePill, ArchetypeRow } from '../components/ArchetypePill';
import { CompareBars } from '../components/Bars';
import { Button } from '../components/Button';
import { Card } from '../components/Card';
import { Chip, ChipRow } from '../components/Chip';
import { Headline, SubCopy } from '../components/Headline';
import { Illustration } from '../components/Illustration';
import { Screen } from '../components/Screen';
import type { Archetype } from '../components/types';
import { Icon } from '../icons';
import { ARCHETYPE_BLURB, type MoneyProps, useMoney } from './shared';

/* One of the four monthly-recap screens, composed from the system and revised per PRD §13.
   Props carry the data; layout is fixed. These are the reference compositions for the rest of the app. */

/** The mascot poster for each Archetype. Prompts: my-money-recap/illustration-prompts.md. */
const ARCHETYPE_ILLUSTRATION: Record<Archetype, string> = {
  squirrel: ILLUSTRATIONS.squirrel,
  hustler: ILLUSTRATIONS.hustler,
  treatYourself: ILLUSTRATIONS.treat,
  slowBurner: ILLUSTRATIONS.slowburner,
};
const ARCHETYPE_TITLE: Record<Archetype, string> = {
  squirrel: 'The Squirrel',
  hustler: 'The Hustler',
  treatYourself: 'Treat-Yourself',
  slowBurner: 'The Slow Burner',
};

export interface RecapPersonalityScreenProps extends MoneyProps {
  childName?: string;
  archetype?: Archetype;
  archetypeLabel?: string;
  blurb?: string;
  saved?: number;
  spent?: number;
  animate?: boolean;
}
/** Screen 4. The shareable Archetype badge. No page dots, so screenshots stay clean. */
export function RecapPersonalityScreen({
  childName = 'Maya',
  archetype = 'squirrel',
  archetypeLabel,
  blurb,
  saved = 640,
  spent = 1760,
  animate = false,
  currency,
  locale,
}: RecapPersonalityScreenProps) {
  const money = useMoney({ currency, locale });
  const all: Archetype[] = ['squirrel', 'hustler', 'treatYourself', 'slowBurner'];
  const title = archetypeLabel ?? ARCHETYPE_TITLE[archetype];
  return (
    <Screen ground="personality">
      <ChipRow>
        <Chip asLabel>Your money personality</Chip>
      </ChipRow>
      <div style={{ marginTop: 16, fontSize: 15, fontWeight: 800, opacity: 0.85 }}>
        This month, {childName}, you were
      </div>
      <Headline size="display" style={{ marginTop: 2 }}>
        {title}
      </Headline>
      <SubCopy style={{ marginTop: 12, opacity: 0.9 }}>{blurb ?? ARCHETYPE_BLURB[archetype]}</SubCopy>
      <Illustration src={ARCHETYPE_ILLUSTRATION[archetype]} width={340} maxHeight={300} />
      <Card underIllustration style={{ gap: 14 }} animateIn={animate}>
        <CompareBars
          animate={animate}
          rows={[
            { label: 'Saved', value: saved, display: money(saved), tone: 'green' },
            { label: 'Spent', value: spent, display: money(spent), tone: 'coral' },
          ]}
        />
        <Button icon={<Icon name="share" />}>Share your {title.replace(/^The /, '')}</Button>
      </Card>
      <div style={{ marginTop: 12, marginLeft: 18 }}>
        <ArchetypeRow>
          {all.map((a) => (
            <ArchetypePill key={a} archetype={a} selected={a === archetype} />
          ))}
        </ArchetypeRow>
      </div>
    </Screen>
  );
}
