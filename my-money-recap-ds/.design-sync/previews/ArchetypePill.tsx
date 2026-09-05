import { ArchetypePill, ArchetypeRow } from '@my-money-recap/ui';

const ground = { background: 'var(--mmr-color-ground-personality)', fontFamily: 'var(--mmr-font-family-text)', padding: 24, width: '100%', maxWidth: 342, boxSizing: 'border-box' as const, borderRadius: 16 };

export function Selected() { return <div style={ground}><ArchetypePill archetype="squirrel" selected /></div>; }
export function Unselected() { return <div style={ground}><ArchetypePill archetype="hustler" /></div>; }
export function Pressed() { return <div style={ground}><ArchetypePill archetype="treatYourself" state="pressed" /></div>; }
export function Disabled() { return <div style={ground}><ArchetypePill archetype="slowBurner" state="disabled" /></div>; }
export function Tappable() { return <div style={ground}><ArchetypePill archetype="squirrel" selected tappable /></div>; }
export function Row() {
  return (
    <div style={ground}>
      <ArchetypeRow>
        <ArchetypePill archetype="squirrel" selected />
        <ArchetypePill archetype="hustler" />
        <ArchetypePill archetype="treatYourself" />
        <ArchetypePill archetype="slowBurner" />
      </ArchetypeRow>
    </div>
  );
}
