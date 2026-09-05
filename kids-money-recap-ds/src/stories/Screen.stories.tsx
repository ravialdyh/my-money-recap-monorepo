import type { Meta, StoryObj } from './_csf';
import { Screen, type ScreenProps } from '../components/Screen';
import { Headline } from '../components/Headline';

const meta: Meta<ScreenProps> = { title: 'Foundations/Screen', component: Screen, parameters: { layout: 'centered' } };
export default meta;
type S = StoryObj<ScreenProps>;
const body = (t: string) => <Headline size="m">{t}</Headline>;
export const Recap: S = { args: { ground: 'recap', children: body('Violet ground') } };
export const Earning: S = { args: { ground: 'earning', children: body('Yellow ground, dark ink') } };
export const Goal: S = { args: { ground: 'goal', children: body('Coral ground') } };
export const Personality: S = { args: { ground: 'personality', children: body('Green ground') } };
