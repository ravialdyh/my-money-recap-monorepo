import { Banner } from '@my-money-recap/ui';

const ground = { background: 'var(--mmr-color-ground-recap)', fontFamily: 'var(--mmr-font-family-text)', padding: 24, width: '100%', maxWidth: 342, boxSizing: 'border-box' as const, borderRadius: 16 };

export function TargetWarning() { return <div style={ground}><Banner>You've passed your saving line. £185 over.</Banner></div>; }
export function Dismissable() { return <div style={ground}><Banner onDismiss={() => {}}>You've passed your saving line. £185 over.</Banner></div>; }
export function Info() { return <div style={ground}><Banner tone="info" onDismiss={() => {}}>Your September recap is ready.</Banner></div>; }

/** D58 / D63: the grey offline notice at the top of Home and Settings. */
export function Offline() { return <div style={ground}><Banner tone="offline">No connection. You can look, but not log.</Banner></div>; }
