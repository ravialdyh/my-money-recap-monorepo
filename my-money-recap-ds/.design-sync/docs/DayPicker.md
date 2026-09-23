---
category: Components
---

The Payday picker: a horizontally scrolling row of 44pt pills, 1–31 followed by `Last day`, with
the selection drawn as the ink pill. It appears twice in the app, onboarding step 3 and the
Settings salary screen, and nowhere else.

**No individual day is ever disabled.** 29, 30 and 31 are offered in every month (PRD D51): the
picker cannot guard the eleven months the Member is not looking at, and disabling them in a February
setup would lock out anyone genuinely paid on the 30th. That is a separate matter from
`state="disabled"`, which dims and inerts the *whole control*: the offline Settings case in
PRD §6.19. There is no per-day disable and there is not meant to be one.

**`nextPaydays` is the answer to "what happens in a short month?"** Pass the three real dates that
follow the current selection (`['30 Nov', '31 Dec', '31 Jan']`) and the component renders
`Next paydays: 30 Nov · 31 Dec · 31 Jan` beneath the pills. **This component does no calendar
arithmetic.** `min(Payday, days in month)` is the caller's job (PRD §8.2); the picker only draws
what it is handed. Pass `null` to remove the line.

**`value` is `number | 'last'`.** The row scrolls to centre the selection when the picker opens, and
deliberately does not re-centre on each tap.

```tsx
const [payday, setPayday] = useState<PaydayValue>(25);
<Card>
  <CardTitle>Payday</CardTitle>
  <DayPicker value={payday} onChange={setPayday} nextPaydays={nextThree(payday)} />
</Card>
```
