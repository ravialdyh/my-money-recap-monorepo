import type { Meta, StoryObj } from './_csf';
import { Chip, ChipRow, type ChipProps } from '../components/Chip';

const meta: Meta<ChipProps> = { title: 'Components/Chip', component: Chip, parameters: { backgrounds: { default: 'recap' } } };
export default meta;
type S = StoryObj<ChipProps>;
export const OnGround: S = { args: { variant: 'onGround', children: 'September 2026' } };
export const Inverse: S = { args: { variant: 'inverse', children: 'Earning' } };
export const Highlight: S = { args: { variant: 'highlight', children: "Maya's recap" } };
export const Pressed: S = { args: { variant: 'onGround', state: 'pressed', children: 'Pressed' } };
export const Disabled: S = { args: { variant: 'onGround', state: 'disabled', children: 'Disabled' } };
export const AsLabel: S = { args: { variant: 'onGround', asLabel: true, children: 'Static label' } };
export const Row: S = { render: () => <ChipRow><Chip>September 2026</Chip><Chip variant="highlight">Maya's recap</Chip></ChipRow> };
export const CategorySelected: S = { args: { tone: 'coral', selected: true, children: 'Snacks' } };
export const CategoryUnselected: S = { args: { tone: 'violet', children: 'Games' } };
