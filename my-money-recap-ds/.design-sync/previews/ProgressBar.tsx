import { Card, CardTitle, HelperText, Icon, ProgressBar } from '@my-money-recap/ui';

const ground = { background: 'var(--mmr-color-ground-goal)', fontFamily: 'var(--mmr-font-family-text)', padding: 24, width: '100%', maxWidth: 342, boxSizing: 'border-box' as const, borderRadius: 16 };

export function Goal() {
  return (
    <div style={ground}>
      <Card>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline' }}>
          <CardTitle>Saving target</CardTitle>
          <span style={{ fontSize: 14, fontWeight: 800, color: 'var(--mmr-color-text-secondary)' }}><span style={{ color: 'var(--mmr-color-text-primary)', fontSize: 18 }}>£640</span> of £480</span>
        </div>
        <div style={{ paddingTop: 14 }}><ProgressBar value={480} max={480} marker={<Icon name="piggy" />} /></div>
        <HelperText>Target met. £160 to spare.</HelperText>
      </Card>
    </div>
  );
}
export function Tones() {
  return (
    <div style={ground}>
      <Card style={{ gap: 12 }}>
        <ProgressBar value={25} max={100} tone="coral" />
        <ProgressBar value={50} max={100} tone="violet" />
        <ProgressBar value={75} max={100} tone="yellow" />
        <ProgressBar value={90} max={100} tone="green" />
        <ProgressBar value={40} max={100} tone="sky" />
        <ProgressBar value={60} max={100} tone="gold" />
        <ProgressBar value={30} max={100} tone="rust" />
        <ProgressBar value={80} max={100} tone="ink" />
      </Card>
    </div>
  );
}
export function Complete() {
  return (
    <div style={ground}>
      <Card>
        <CardTitle>Saving target, missed</CardTitle>
        <div style={{ paddingTop: 14 }}><ProgressBar value={310} max={480} tone="coral" marker={<Icon name="piggy" />} /></div>
        <HelperText>£170 short of your target.</HelperText>
      </Card>
    </div>
  );
}
