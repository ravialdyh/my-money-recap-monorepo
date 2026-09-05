import { Card, TextField } from '@kids-money-recap/ui';

const ground = { background: 'var(--kmr-color-ground-recap)', fontFamily: 'var(--kmr-font-family-text)', padding: 24, width: '100%', maxWidth: 342, boxSizing: 'border-box' as const, borderRadius: 16 };

export function Default() { return <div style={ground}><Card><TextField label="Email address" placeholder="you@example.com" /></Card></div>; }
export function Filled() { return <div style={ground}><Card><TextField label="First name" defaultValue="Maya" /></Card></div>; }
export function Focus() { return <div style={ground}><Card><TextField label="Email address" defaultValue="maya@example.com" state="focus" /></Card></div>; }
export function Invalid() { return <div style={ground}><Card><TextField label="Email address" defaultValue="maya@example" state="invalid" helper="That doesn't look like an email address." /></Card></div>; }
export function Disabled() { return <div style={ground}><Card><TextField label="Email address" defaultValue="maya@example.com" state="disabled" /></Card></div>; }
export function WithCounter() { return <div style={ground}><Card><TextField label="Note" defaultValue="Concert ticket" maxLength={80} suffix="14/80" helper="Optional, 80 characters." /></Card></div>; }
