import type { Meta, StoryObj } from './_csf';
import { TextField, type TextFieldProps } from '../components/TextField';
import { AmountField } from '../components/AmountField';
import { Slider } from '../components/Slider';
import { Toggle } from '../components/Toggle';

const meta: Meta<TextFieldProps> = { title: 'Components/Forms', component: TextField, parameters: { backgrounds: { default: 'white' } } };
export default meta;
type S = StoryObj<TextFieldProps>;
export const Field: S = { args: { label: 'Email address', placeholder: 'you@example.com' } };
export const FieldFocus: S = { args: { label: 'Email address', defaultValue: 'maya@example.com', state: 'focus' } };
export const FieldInvalid: S = { args: { label: 'Email address', defaultValue: 'maya@example', state: 'invalid', helper: "That doesn't look like an email address." } };
export const Amount: S = { render: () => <AmountField label="Your monthly salary" currency="GBP" valueMinor={240000} /> };
export const AmountRupiah: S = { render: () => <AmountField label="Your monthly salary" currency="IDR" valueMinor={3000000} /> };
export const TargetSlider: S = { render: () => <div style={{ width: 300 }}><Slider value={20} helper="£480 of £2,400 each month" aria-label="Saving target" /></div> };
export const Switches: S = { render: () => (
  <div style={{ display: 'grid', gap: 12 }}>
    <Toggle checked aria-label="Recap ready" />
    <Toggle checked={false} aria-label="Daily reminder" />
    <Toggle checked disabled aria-label="Target warning" />
  </div>
) };
