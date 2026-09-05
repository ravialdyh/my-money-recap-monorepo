import type { Meta, StoryObj } from './_csf';
import { StatHero, type StatHeroProps } from '../components/StatHero';
import { StatChip, StatChipRow } from '../components/StatChip';

const meta: Meta<StatHeroProps> = { title: 'Components/Stat', component: StatHero, parameters: { backgrounds: { default: 'white' } } };
export default meta;
type S = StoryObj<StatHeroProps>;
export const Hero: S = { args: { eyebrow: 'Money that came your way', value: 36 } };
export const HeroCountUp: S = { args: { eyebrow: 'Money that came your way', value: 36, animate: true } };
export const Chips: S = { render: () => (
  <StatChipRow>
    <StatChip tone="earned">Money in £2,400</StatChip>
    <StatChip tone="saved">Saved £640</StatChip>
    <StatChip tone="spent">Spent £1,760</StatChip>
  </StatChipRow>
) };
export const ChipStates: S = { render: () => (
  <StatChipRow>
    <StatChip tone="saved">Default</StatChip>
    <StatChip tone="saved" state="pressed">Pressed</StatChip>
    <StatChip tone="saved" state="disabled">Disabled</StatChip>
  </StatChipRow>
) };
