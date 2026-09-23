---
category: Components
---

The empty-state card. Use it wherever the Member has not given the app something yet and the
screen would otherwise show a hole: the Home hero with no Salary, the Home target card with no
Saving Target, onboarding step 4 when Salary was skipped, Recap screen 4 with no Archetype, a
Settings row that reads "Not set".

**The whole card is one tap target, not the chevron.** `PromptCard` *is* the `<button>`: do not
wrap it in one, and do not put a control inside it. Because a `<button>` may not contain an `<h2>`
or a `<p>`, it renders `Eyebrow`, `CardTitle` and `HelperText` in their `as="span"` form; the drawn
result is identical to the same three inside a plain `Card`.

**`height` is the point of the component.** A prompt stands in for a filled card, so it holds that
card's height and the layout does not jump when the Member fills it in. The height is measured from
the card it replaces as drawn on this screen, never assumed (PRD D71): the pound Home hero is 191pt,
the rupiah one 233pt because its stat chips wrap (D45), the Home target card 136pt. Omit `height`
where there is nothing to stand in for (a Settings row, Recap screen 4) and the card hugs its
content.

**`eyebrow` and `helper` take `null` to remove them**, and `null` is the default, so a card with
neither is `<PromptCard title="…" />` and a card with both names them explicitly.

```tsx
// Home, no Salary set (PRD D54): the hero keeps Spent, the prompt sits below it at the hero's measured height.
<Card><StatHero eyebrow="Spent so far" value={1760} currency="GBP" /></Card>
<PromptCard
  height={heroHeight}
  title="Set your monthly salary"
  helper="We'll show what you kept once we know what came in."
  onClick={openSalary}
/>
```
