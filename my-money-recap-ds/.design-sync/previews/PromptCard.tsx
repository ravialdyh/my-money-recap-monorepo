import { Card, CardTitle, HelperText, PromptCard, StatHero, StatChip, StatChipRow } from '@my-money-recap/ui';

const ground = { background: 'var(--mmr-color-ground-recap)', fontFamily: 'var(--mmr-font-family-text)', padding: 24, width: '100%', maxWidth: 342, boxSizing: 'border-box' as const, borderRadius: 16 };
const stack = { display: 'flex', flexDirection: 'column' as const, gap: 14 };

/** Home, no Salary set (D54): the hero keeps Spent and the prompt sits below it. */
export function HomeHeroPrompt() {
  return (
    <div style={ground}><div style={stack}>
      <Card><StatHero eyebrow="Spent so far" value={1760} format={(n) => `£${n.toLocaleString('en-GB')}`} />
        <StatChipRow><StatChip tone="earned">Money in £0</StatChip></StatChipRow>
      </Card>
      <PromptCard size="hero" title="Set your monthly salary" helper="We'll show what you kept once we know what came in." />
    </div></div>
  );
}
/** Home, no Saving Target set: the 136pt target card becomes the prompt, so nothing below it moves. */
export function HomeTargetPrompt() {
  return (
    <div style={ground}><div style={stack}>
      <PromptCard size="target" title="Set a saving target" helper="Pick a share of your salary to keep each month." />
      <Card style={{ height: 136 }}>
        <CardTitle>Saving target</CardTitle>
        <HelperText>The filled card it stands in for, at the same 136pt.</HelperText>
      </Card>
    </div></div>
  );
}
/** Onboarding step 4 with Salary skipped (D44, A7). Tapping returns to step 3. */
export function OnboardingSalaryFirst() {
  return (
    <div style={ground}>
      <PromptCard size="target" eyebrow="Step 4 of 4" title="Set a salary first" helper="A target is a share of your salary, so we need the salary first." />
    </div>
  );
}
/** Recap screen 4 with no Archetype (C4): the prompt stands in for the mascot. */
export function RecapNoArchetype() {
  return (
    <div style={{ ...ground, background: 'var(--mmr-color-ground-personality)' }}>
      <PromptCard size="hero" title="Set a saving target to get an archetype" helper="Your archetype comes from what you kept against what you meant to keep." />
    </div>
  );
}
/** A Settings row: `auto` hugs its content, with no fixed height to hold. */
export function SettingsRow() {
  return <div style={ground}><PromptCard title="Salary and payday" helper="Not set" /></div>;
}
/** `eyebrow={null}` and `helper={null}` remove both: the card is the title and the chevron alone. */
export function TitleOnly() {
  return (
    <div style={ground}><div style={stack}>
      <PromptCard eyebrow="Step 4 of 4" title="Set a salary first" helper="Both slots filled, for comparison." />
      <PromptCard eyebrow={null} title="Set a salary first" helper={null} />
    </div></div>
  );
}
export function Pressed() { return <div style={ground}><PromptCard size="target" title="Set a saving target" helper="Pick a share of your salary to keep each month." state="pressed" /></div>; }
export function Disabled() { return <div style={ground}><PromptCard size="target" title="Set a saving target" helper="Offline — you can look, but not log." state="disabled" /></div>; }
