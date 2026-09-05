import type { Meta, StoryObj } from './_csf';
import { Card, CardTitle, type CardProps } from '../components/Card';
import { Eyebrow } from '../components/Headline';

const meta: Meta<CardProps> = { title: 'Components/Card', component: Card, parameters: { backgrounds: { default: 'recap' } } };
export default meta;
type S = StoryObj<CardProps>;
const inner = <><Eyebrow>Where £11 went</Eyebrow><CardTitle>Mostly snacks</CardTitle></>;
export const Surface: S = { args: { variant: 'surface', children: inner } };
export const Accent: S = { args: { variant: 'accent', children: <CardTitle>Longest no-spend streak</CardTitle> } };
export const UnderIllustration: S = { args: { variant: 'surface', underIllustration: true, children: inner } };
export const AnimateIn: S = { args: { variant: 'surface', animateIn: true, children: inner } };
