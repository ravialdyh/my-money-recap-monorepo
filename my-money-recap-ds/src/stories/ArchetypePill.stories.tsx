import type { Meta, StoryObj } from './_csf';
import { ArchetypePill, ArchetypeRow, type ArchetypePillProps } from '../components/ArchetypePill';

const meta: Meta<ArchetypePillProps> = { title: 'Components/ArchetypePill', component: ArchetypePill, parameters: { backgrounds: { default: 'personality' } } };
export default meta;
type S = StoryObj<ArchetypePillProps>;
export const Selected: S = { args: { archetype: 'squirrel', selected: true } };
export const Unselected: S = { args: { archetype: 'hustler' } };
export const Pressed: S = { args: { archetype: 'treatYourself', state: 'pressed' } };
export const Disabled: S = { args: { archetype: 'slowBurner', state: 'disabled' } };
export const Row: S = { render: () => (
  <ArchetypeRow>
    <ArchetypePill archetype="squirrel" selected />
    <ArchetypePill archetype="hustler" />
    <ArchetypePill archetype="treatYourself" />
    <ArchetypePill archetype="slowBurner" />
  </ArchetypeRow>
) };
