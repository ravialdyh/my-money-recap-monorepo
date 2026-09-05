import type { Meta, StoryObj } from './_csf';
import { Headline, SubCopy, HelperText, Eyebrow, type HeadlineProps } from '../components/Headline';

const meta: Meta<HeadlineProps> = { title: 'Foundations/Type', component: Headline };
export default meta;
type S = StoryObj<HeadlineProps>;
export const HeadlineL: S = { args: { size: 'l', children: 'Your September in money' } };
export const HeadlineM: S = { args: { size: 'm', children: '133% of your saving target' } };
export const HeadlineCompact: S = { args: { size: 'compact', children: "You kept £640 of £2,400. That's 133% of your target." } };
export const Display: S = { args: { size: 'display', children: 'The Squirrel' } };
export const Copy: S = { render: () => (
  <div style={{ display: 'grid', gap: 12, color: '#1E1B3A' }}>
    <SubCopy>About 3 weeks to go at this rate.</SubCopy>
    <HelperText>Swipe to see where it all went.</HelperText>
    <Eyebrow>Money that came your way</Eyebrow>
    <Eyebrow tone="coral">The splurge</Eyebrow>
  </div>
) };
