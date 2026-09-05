import type { Meta, StoryObj } from './_csf';
import { DataRow, DataRowList, type DataRowProps } from '../components/DataRow';
import { Icon } from '../icons';

const meta: Meta<DataRowProps> = { title: 'Components/DataRow', component: DataRow, parameters: { backgrounds: { default: 'white' } } };
export default meta;
type S = StoryObj<DataRowProps>;
export const Default: S = { args: { icon: <Icon name="paw" />, iconTone: 'violet', label: '8 dog walks', amount: '£16' } };
export const Pressed: S = { args: { icon: <Icon name="droplet" />, iconTone: 'sky', label: '1 car wash', amount: '£5', state: 'pressed' } };
export const Disabled: S = { args: { icon: <Icon name="lemon" />, iconTone: 'coral', label: 'Lemonade stand', amount: '£3', state: 'disabled' } };
export const List: S = { render: () => (
  <DataRowList>
    <DataRow icon={<Icon name="paw" />} iconTone="violet" label="8 dog walks" amount="£16" />
    <DataRow icon={<Icon name="droplet" />} iconTone="sky" label="1 car wash" amount="£5" />
    <DataRow icon={<Icon name="lemon" />} iconTone="coral" label="Lemonade stand" amount="£3" />
  </DataRowList>
) };
