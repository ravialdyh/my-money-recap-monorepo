import type { Meta, StoryObj } from './_csf';
import { DataRow, DataRowList, type DataRowProps } from '../components/DataRow';
import { Icon } from '../icons';

const meta: Meta<DataRowProps> = { title: 'Components/DataRow', component: DataRow, parameters: { backgrounds: { default: 'white' } } };
export default meta;
type S = StoryObj<DataRowProps>;
export const Default: S = { args: { icon: <Icon name="paw" />, iconTone: 'violet', label: 'September 2026', amount: '£16' } };
export const Pressed: S = { args: { icon: <Icon name="droplet" />, iconTone: 'sky', label: 'August 2026', amount: '£5', state: 'pressed' } };
export const Disabled: S = { args: { icon: <Icon name="lemon" />, iconTone: 'coral', label: 'July 2026', amount: '£3', state: 'disabled' } };
export const List: S = { render: () => (
  <DataRowList>
    <DataRow icon={<Icon name="paw" />} iconTone="violet" label="September 2026" amount="Squirrel" />
    <DataRow icon={<Icon name="droplet" />} iconTone="sky" label="August 2026" amount="Hustler" />
    <DataRow icon={<Icon name="lemon" />} iconTone="coral" label="July 2026" amount="Slow Burner" />
  </DataRowList>
) };
