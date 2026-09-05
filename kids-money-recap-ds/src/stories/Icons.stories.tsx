import type { Meta, StoryObj } from './_csf';
import { Icon, ICON_NAMES, type IconProps } from '../icons';
import { IconWell } from '../components/DataRow';

const meta: Meta<IconProps> = { title: 'Foundations/Icon', component: Icon, parameters: { backgrounds: { default: 'white' } } };
export default meta;
type S = StoryObj<IconProps>;
export const Paw: S = { args: { name: 'paw' } };
export const Droplet: S = { args: { name: 'droplet' } };
export const Lemon: S = { args: { name: 'lemon' } };
export const Share: S = { args: { name: 'share', style: { color: '#1E1B3A' } } };
export const Sparkle: S = { args: { name: 'sparkle', size: 24, style: { background: '#3DAA6A', borderRadius: 4 } } };
export const Piggy: S = { args: { name: 'piggy' } };
export const Categories: S = { render: () => (
  <div style={{ display: 'flex', gap: 12, alignItems: 'center', flexWrap: 'wrap' }}>
    <IconWell tone="coral"><Icon name="snacks" /></IconWell>
    <IconWell tone="violet"><Icon name="games" /></IconWell>
    <IconWell tone="yellow"><Icon name="toys" /></IconWell>
    <IconWell tone="sky"><Icon name="rides" /></IconWell>
    <IconWell tone="green"><Icon name="clothes" /></IconWell>
    <IconWell tone="ink"><Icon name="bills" /></IconWell>
    <IconWell tone="gold"><Icon name="gifts" /></IconWell>
    <IconWell tone="rust"><Icon name="other" /></IconWell>
  </div>
) };
export const UiGlyphs: S = { render: () => (
  <div style={{ display: 'flex', gap: 12, color: '#1E1B3A' }}>
    <Icon name="settings" /><Icon name="chevron" /><Icon name="plus" />
  </div>
) };
export const AllInWells: S = { render: () => (
  <div style={{ display: 'flex', gap: 12, alignItems: 'center' }}>
    <IconWell tone="violet"><Icon name="paw" /></IconWell>
    <IconWell tone="sky"><Icon name="droplet" /></IconWell>
    <IconWell tone="coral"><Icon name="lemon" /></IconWell>
    <span style={{ color: '#1E1B3A', display: 'inline-flex' }}><Icon name="share" /></span>
    <span style={{ display: 'inline-flex', background: '#3DAA6A', borderRadius: 4 }}><Icon name="sparkle" size={20} /></span>
    <Icon name="piggy" />
    <span style={{ fontSize: 12, color: '#6B6890' }}>{ICON_NAMES.join(' · ')}</span>
  </div>
) };
