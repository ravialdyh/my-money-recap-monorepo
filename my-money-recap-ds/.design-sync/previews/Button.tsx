import { Button, Card, Icon } from '@my-money-recap/ui';

const ground = { background: 'var(--mmr-color-ground-personality)', fontFamily: 'var(--mmr-font-family-text)', padding: 24, width: '100%', maxWidth: 342, boxSizing: 'border-box' as const, borderRadius: 16 };

export function Primary() { return <div style={ground}><Card><Button icon={<Icon name="share" />}>Share your Squirrel</Button></Card></div>; }
export function Pressed() { return <div style={ground}><Card><Button icon={<Icon name="share" />} state="pressed">Share your Squirrel</Button></Card></div>; }
export function Disabled() { return <div style={ground}><Card><Button icon={<Icon name="share" />} state="disabled">Share your Squirrel</Button></Card></div>; }
export function Secondary() { return <div style={ground}><Card><Button variant="secondary">See last month</Button></Card></div>; }
export function SecondaryDisabled() { return <div style={ground}><Card><Button variant="secondary" state="disabled">See last month</Button></Card></div>; }
export function Pair() {
  return <div style={ground}><Card><Button>Convert</Button><Button variant="secondary">Cancel</Button></Card></div>;
}
export function NoIcon() { return <div style={ground}><Card><Button>Continue</Button></Card></div>; }
