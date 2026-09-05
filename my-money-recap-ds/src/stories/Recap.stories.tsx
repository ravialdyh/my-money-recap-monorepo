import type { Meta, StoryObj } from './_csf';
import { RecapCoverScreen, RecapEarningScreen, RecapGoalScreen, RecapPersonalityScreen } from '../screens/RecapScreens';

const meta: Meta = { title: 'Screens/Monthly recap', parameters: { layout: 'centered' } };
export default meta;
export const Cover: StoryObj = { render: () => <RecapCoverScreen /> };
export const Earning: StoryObj = { render: () => <RecapEarningScreen /> };
export const GoalAndSpending: StoryObj = { render: () => <RecapGoalScreen /> };
export const Personality: StoryObj = { render: () => <RecapPersonalityScreen /> };
export const CoverAnimated: StoryObj = { render: () => <RecapCoverScreen animate /> };
