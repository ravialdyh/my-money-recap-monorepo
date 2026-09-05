import type { Meta, StoryObj } from './_csf';
import { Button, type ButtonProps } from '../components/Button';
import { Icon } from '../icons';

const meta: Meta<ButtonProps> = { title: 'Components/Button', component: Button, parameters: { backgrounds: { default: 'white' } } };
export default meta;
type S = StoryObj<ButtonProps>;
export const Primary: S = { args: { icon: <Icon name="share" />, children: 'Share your Squirrel' } };
export const Pressed: S = { args: { icon: <Icon name="share" />, state: 'pressed', children: 'Share your Squirrel' } };
export const Disabled: S = { args: { icon: <Icon name="share" />, state: 'disabled', children: 'Share your Squirrel' } };
export const Secondary: S = { args: { variant: 'secondary', children: 'See last month' } };
export const SecondaryDisabled: S = { args: { variant: 'secondary', state: 'disabled', children: 'See last month' } };
export const NoIcon: S = { args: { children: 'Continue' } };
