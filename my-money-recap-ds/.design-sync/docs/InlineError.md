---
category: Components
---

The app's only error surface. `Toast` was not built (PRD D67), so every failure the Member can see
is an `InlineError`: sign-in cancelled or failed, an expired or reused magic link, a spend that
would not save, a currency rate that would not fetch, an export with nowhere to send.

**It pushes content down and never overlays it.** That is the whole reason it beat the toast
(PRD §6.1, §6.19): a block in the flow cannot cover the button the Member needs next, and it stays
on screen until the state that caused it changes. Put it where the failure happened, inside the
card, above the action, never at the top of the screen and never absolutely positioned.

**`action` is the trailing control**: `Try again` on a failed rate fetch, a link to Profile on an
undeliverable address. It takes `null` to remove it, and `null` is the default, so a plain message
is just `<InlineError>…</InlineError>`.

Coral-100 fill, a coral-500 dot, coral text in Nunito 800 13. On a surface that is already coral, the restore
sheet's (SCREEN-STATES D11b), the fill is white instead; the app's `InlineError` takes it as `well`. It carries `role="alert"`, so do not
add another live region around it.

```tsx
<Card>
  <InlineError action={<a href="#retry">Try again</a>}>We couldn't get today's rate.</InlineError>
  <Button>Convert</Button>
</Card>
```
