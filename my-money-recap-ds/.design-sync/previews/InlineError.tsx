import { Button, Card, InlineError, TextField } from '@my-money-recap/ui';

const ground = { background: 'var(--mmr-color-ground-recap)', fontFamily: 'var(--mmr-font-family-text)', padding: 24, width: '100%', maxWidth: 342, boxSizing: 'border-box' as const, borderRadius: 16 };
const stack = { display: 'flex', flexDirection: 'column' as const, gap: 14 };

/** Sign in, cancelled or failed (A1). The block sits inside the card, above the buttons. */
export function SignIn() {
  return (
    <div style={ground}><Card><div style={stack}>
      <InlineError>We couldn't finish signing you in. Try again.</InlineError>
      <Button>Continue with email</Button>
    </div></Card></div>
  );
}
/** Email entry, expired link (A3). One line, no action. */
export function LinkExpired() {
  return <div style={ground}><Card><InlineError>That link has expired. Links work for 15 minutes.</InlineError></Card></div>;
}
/** Currency rate fetch failed (D6): the action slot carries `Try again`. */
export function WithAction() {
  return (
    <div style={ground}><Card>
      <InlineError action={<a href="#try-again">Try again</a>}>We couldn't get today's rate.</InlineError>
    </Card></div>
  );
}
/** `action={null}` removes the trailing control: the same message, nothing on the right. */
export function NoAction() {
  return (
    <div style={ground}><Card><div style={stack}>
      <InlineError action={<a href="#try-again">Try again</a>}>We couldn't get today's rate.</InlineError>
      <InlineError action={null}>We couldn't get today's rate.</InlineError>
    </div></Card></div>
  );
}
/** Log a spend, write failed (B2): the error pushes Save down, it never covers it (§6.19). */
export function PushesContentDown() {
  return (
    <div style={ground}><Card><div style={stack}>
      <TextField label="Note" defaultValue="Concert ticket" />
      <InlineError>We couldn't save that spend. Your amount and category are still here.</InlineError>
      <Button>Save</Button>
    </div></Card></div>
  );
}
/** Export, no deliverable address (D9). Two lines, to show the dot staying on the first. */
export function TwoLines() {
  return (
    <div style={ground}><Card>
      <InlineError action={<a href="#profile">Profile</a>}>We can't reach maya@privaterelay.appleid.com. Add an address we can send the CSV to.</InlineError>
    </Card></div>
  );
}
