import { Card, CardTitle, Slider } from '@my-money-recap/ui';

const ground = { background: 'var(--mmr-color-ground-recap)', fontFamily: 'var(--mmr-font-family-text)', padding: 24, width: '100%', maxWidth: 342, boxSizing: 'border-box' as const, borderRadius: 16 };

export function Default() { return <div style={ground}><Card><CardTitle>20%</CardTitle><Slider value={20} helper="£480 of £2,400 each month" aria-label="Saving target" /></Card></div>; }
export function Zero() { return <div style={ground}><Card><CardTitle>0%</CardTitle><Slider value={0} helper="Keep nothing set aside." aria-label="Saving target" /></Card></div>; }
export function Half() { return <div style={ground}><Card><CardTitle>50%</CardTitle><Slider value={50} helper="£1,200 of £2,400 each month" aria-label="Saving target" /></Card></div>; }
export function Full() { return <div style={ground}><Card><CardTitle>100%</CardTitle><Slider value={100} helper="£2,400 of £2,400 each month" aria-label="Saving target" /></Card></div>; }
export function NoSalary() { return <div style={ground}><Card><CardTitle>20%</CardTitle><Slider value={20} aria-label="Saving target" /></Card></div>; }
export function Disabled() { return <div style={ground}><Card><CardTitle>20%</CardTitle><Slider value={20} disabled helper="Set a salary first." aria-label="Saving target" /></Card></div>; }
