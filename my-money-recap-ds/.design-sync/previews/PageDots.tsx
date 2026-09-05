import { PageDots } from '@my-money-recap/ui';

const ground = (g: string) => ({ background: `var(--mmr-color-ground-${g})`, fontFamily: 'var(--mmr-font-family-text)', padding: 24, width: '100%', maxWidth: 342, boxSizing: 'border-box' as const, borderRadius: 16 });

export function OnGround() { return <div style={ground('recap')}><PageDots count={4} active={0} /></div>; }
export function OnLight() { return <div style={ground('earning')}><PageDots count={4} active={1} onLight /></div>; }
export function LastPage() { return <div style={ground('goal')}><PageDots count={4} active={3} /></div>; }
