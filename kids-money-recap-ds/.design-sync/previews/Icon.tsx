import { Icon, IconWell, ICON_NAMES } from '@kids-money-recap/ui';

const row = { display: 'flex', gap: 16, alignItems: 'center', padding: 8 };
const wrap = { display: 'flex', gap: 12, alignItems: 'center', flexWrap: 'wrap' as const, padding: 8 };

export function Paw() { return <div style={row}><Icon name="paw" /><Icon name="paw" size={32} /></div>; }
export function Droplet() { return <div style={row}><Icon name="droplet" /><Icon name="droplet" size={32} /></div>; }
export function Lemon() { return <div style={row}><Icon name="lemon" /><Icon name="lemon" size={32} /></div>; }
export function Share() { return <div style={{ ...row, color: 'var(--kmr-color-text-primary)' }}><Icon name="share" /><Icon name="share" size={32} /></div>; }
export function Sparkle() {
  return (
    <div style={row}>
      <span style={{ display: 'inline-flex', background: 'var(--kmr-color-ground-personality)', borderRadius: 4 }}><Icon name="sparkle" size={24} /></span>
      <span style={{ display: 'inline-flex', background: 'var(--kmr-color-ground-recap)', borderRadius: 8, padding: 4 }}><Icon name="sparkle" size={40} /></span>
    </div>
  );
}
export function Markers() { return <div style={row}><Icon name="skateboard" /><Icon name="piggy" /></div>; }
/** The eight Spend Categories, each in its own well tone. */
export function Categories() {
  return (
    <div style={wrap}>
      <IconWell tone="coral"><Icon name="snacks" /></IconWell>
      <IconWell tone="violet"><Icon name="games" /></IconWell>
      <IconWell tone="yellow"><Icon name="toys" /></IconWell>
      <IconWell tone="sky"><Icon name="rides" /></IconWell>
      <IconWell tone="green"><Icon name="clothes" /></IconWell>
      <IconWell tone="ink"><Icon name="bills" /></IconWell>
      <IconWell tone="gold"><Icon name="gifts" /></IconWell>
      <IconWell tone="rust"><Icon name="other" /></IconWell>
    </div>
  );
}
export function UiGlyphs() {
  return (
    <div style={{ ...row, color: 'var(--kmr-color-text-primary)' }}>
      <Icon name="settings" /><Icon name="chevron" /><Icon name="plus" /><Icon name="share" />
    </div>
  );
}
export function All() {
  return (
    <div style={wrap}>
      <IconWell tone="violet"><Icon name="paw" /></IconWell>
      <IconWell tone="sky"><Icon name="droplet" /></IconWell>
      <IconWell tone="coral"><Icon name="lemon" /></IconWell>
      <IconWell tone="coral"><Icon name="snacks" /></IconWell>
      <IconWell tone="violet"><Icon name="games" /></IconWell>
      <IconWell tone="yellow"><Icon name="toys" /></IconWell>
      <IconWell tone="sky"><Icon name="rides" /></IconWell>
      <IconWell tone="green"><Icon name="clothes" /></IconWell>
      <IconWell tone="ink"><Icon name="bills" /></IconWell>
      <IconWell tone="gold"><Icon name="gifts" /></IconWell>
      <IconWell tone="rust"><Icon name="other" /></IconWell>
      <span style={{ color: 'var(--kmr-color-text-primary)', display: 'inline-flex', gap: 8 }}>
        <Icon name="share" /><Icon name="settings" /><Icon name="chevron" /><Icon name="plus" />
      </span>
      <Icon name="piggy" />
      <span style={{ fontSize: 12, fontFamily: 'var(--kmr-font-family-text)', color: 'var(--kmr-color-text-secondary)' }}>{ICON_NAMES.join(' · ')}</span>
    </div>
  );
}
