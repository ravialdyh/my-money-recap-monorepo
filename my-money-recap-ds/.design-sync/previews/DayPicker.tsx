import { Card, CardTitle, DayPicker } from '@my-money-recap/ui';

const ground = { background: 'var(--mmr-color-ground-recap)', fontFamily: 'var(--mmr-font-family-text)', padding: 24, width: '100%', maxWidth: 342, boxSizing: 'border-box' as const, borderRadius: 16 };

/** Onboarding step 3, the default Payday of 25 (PRD §6.6). */
export function Default() {
  return (
    <div style={ground}><Card>
      <CardTitle>Payday</CardTitle>
      <DayPicker value={25} nextPaydays={['25 Sep', '25 Oct', '25 Nov']} />
    </Card></div>
  );
}
/** Payday 31, the case D51 exists for: the three dates say what a short month does. */
export function ThirtyFirst() {
  return (
    <div style={ground}><Card>
      <CardTitle>Payday</CardTitle>
      <DayPicker value={31} nextPaydays={['30 Nov', '31 Dec', '31 Jan']} />
    </Card></div>
  );
}
/** `Last day` is the 32nd pill, and the only one wider than 44pt. */
export function LastDay() {
  return (
    <div style={ground}><Card>
      <CardTitle>Payday</CardTitle>
      <DayPicker value="last" nextPaydays={['30 Nov', '31 Dec', '31 Jan']} />
    </Card></div>
  );
}
/** Day 1, to show the row scrolled to its start rather than centred mid-row. */
export function FirstOfTheMonth() {
  return (
    <div style={ground}><Card>
      <CardTitle>Payday</CardTitle>
      <DayPicker value={1} nextPaydays={['1 Oct', '1 Nov', '1 Dec']} />
    </Card></div>
  );
}
/** `nextPaydays={null}` removes the helper line entirely — pills only, nothing below them. */
export function NoHelperLine() {
  return (
    <div style={ground}><Card>
      <CardTitle>Payday</CardTitle>
      <DayPicker value={25} nextPaydays={null} />
    </Card></div>
  );
}
/** Settings, offline: every row that writes sits at 40% (PRD §6.19). */
export function Disabled() {
  return (
    <div style={ground}><Card>
      <CardTitle>Payday</CardTitle>
      <DayPicker value={25} nextPaydays={['25 Sep', '25 Oct', '25 Nov']} state="disabled" />
    </Card></div>
  );
}
