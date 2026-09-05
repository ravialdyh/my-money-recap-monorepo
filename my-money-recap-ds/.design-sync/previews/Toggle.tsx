import { Card, Toggle } from '@my-money-recap/ui';

const ground = { background: 'var(--mmr-color-ground-recap)', fontFamily: 'var(--mmr-font-family-text)', padding: 24, width: '100%', maxWidth: 342, boxSizing: 'border-box' as const, borderRadius: 16 };
const row = { display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 12, fontSize: 14, fontWeight: 800 };

export function On() { return <div style={ground}><Card><div style={row}><span>Recap ready</span><Toggle checked aria-label="Recap ready" /></div></Card></div>; }
export function Off() { return <div style={ground}><Card><div style={row}><span>Daily reminder</span><Toggle checked={false} aria-label="Daily reminder" /></div></Card></div>; }
export function Disabled() { return <div style={ground}><Card><div style={row}><span>Target warning</span><Toggle checked disabled aria-label="Target warning" /></div></Card></div>; }
export function Group() {
  return (
    <div style={ground}><Card>
      <div style={row}><span>Recap ready</span><Toggle checked aria-label="Recap ready" /></div>
      <div style={row}><span>Daily reminder</span><Toggle checked={false} aria-label="Daily reminder" /></div>
      <div style={row}><span>Target warning</span><Toggle checked aria-label="Target warning" /></div>
    </Card></div>
  );
}
