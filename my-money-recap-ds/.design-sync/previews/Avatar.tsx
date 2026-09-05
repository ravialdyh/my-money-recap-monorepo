import { Avatar, ILLUSTRATIONS } from '@my-money-recap/ui';

const ground = { background: 'var(--mmr-color-ground-recap)', fontFamily: 'var(--mmr-font-family-text)', padding: 24, width: '100%', maxWidth: 342, boxSizing: 'border-box' as const, borderRadius: 16 };

export function Greeting() {
  return (
    <div style={{ ...ground, display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
      <span style={{ fontFamily: 'var(--mmr-font-family-display)', fontSize: 40, fontWeight: 700, color: '#fff' }}>Hi, Maya</span>
      <Avatar src={ILLUSTRATIONS['avatar-otter']} label="Otter" />
    </div>
  );
}
export function Small() { return <div style={ground}><Avatar src={ILLUSTRATIONS['avatar-fox']} label="Fox" size={40} /></div>; }
