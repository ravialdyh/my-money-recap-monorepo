import type { Meta, StoryObj } from './_csf';
import { PageDots, type PageDotsProps } from '../components/PageDots';

const meta: Meta<PageDotsProps> = { title: 'Components/PageDots', component: PageDots, parameters: { backgrounds: { default: 'recap' } } };
export default meta;
type S = StoryObj<PageDotsProps>;
export const OnGround: S = { args: { count: 4, active: 0 } };
export const OnLight: S = { args: { count: 4, active: 1, onLight: true }, parameters: { backgrounds: { default: 'earning' } } };
