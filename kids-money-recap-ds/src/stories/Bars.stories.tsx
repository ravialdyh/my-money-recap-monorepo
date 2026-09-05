import type { Meta, StoryObj } from './_csf';
import { ProgressBar, CategoryBar, CompareBars, type ProgressBarProps } from '../components/Bars';
import { Icon } from '../icons';

const meta: Meta<ProgressBarProps> = { title: 'Components/Bars', component: ProgressBar, parameters: { backgrounds: { default: 'white' } } };
export default meta;
type S = StoryObj<ProgressBarProps>;
export const Goal: S = { render: () => <div style={{ paddingTop: 32, width: 300 }}><ProgressBar value={640} max={480} marker={<Icon name="piggy" />} /></div> };
export const GoalAnimated: S = { render: () => <div style={{ paddingTop: 32, width: 300 }}><ProgressBar value={640} max={480} animate marker={<Icon name="piggy" />} /></div> };
export const Categories: S = { render: () => (
  <div style={{ display: 'grid', gap: 10, width: 220 }}>
    <CategoryBar label="Snacks" percent={45} tone="coral" />
    <CategoryBar label="Games" percent={36} tone="violet" />
    <CategoryBar label="Toys" percent={19} tone="yellow" />
    <CategoryBar label="Rides" percent={26} tone="sky" />
    <CategoryBar label="Clothes" percent={10} tone="green" />
    <CategoryBar label="Bills" percent={35} tone="ink" />
    <CategoryBar label="Gifts" percent={7} tone="gold" />
    <CategoryBar label="Other" percent={1} tone="rust" />
  </div>
) };
export const Compare: S = { render: () => (
  <div style={{ width: 300 }}>
    <CompareBars rows={[{ label: 'Saved', value: 25, display: '£25', tone: 'green' }, { label: 'Spent', value: 11, display: '£11', tone: 'coral' }]} />
  </div>
) };
