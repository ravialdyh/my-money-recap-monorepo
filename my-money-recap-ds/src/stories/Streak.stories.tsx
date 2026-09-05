import type { Meta, StoryObj } from './_csf';
import { StreakBadge, StreakDots, type StreakBadgeProps } from '../components/Streak';

const meta: Meta<StreakBadgeProps> = { title: 'Components/Streak', component: StreakBadge, parameters: { backgrounds: { default: 'violet' } } };
export default meta;
type S = StoryObj<StreakBadgeProps>;
export const Badge: S = { args: { value: 9, unit: 'days' } };
export const Dots: S = { render: () => <div style={{ width: 300 }}><StreakDots total={30} activeRange={[12, 20]} /></div> };
