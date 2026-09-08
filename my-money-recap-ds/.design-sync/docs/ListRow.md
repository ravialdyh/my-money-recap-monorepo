---
category: Components
---

Settings and Currency list rows: a `DataRow` without the `IconWell`, and a chevron or a violet check
where the amount would be.

**`ListRow` is the currency-list row — there is no `SelectRow`.** PRD §13 lists a fourth new
component described as "`ListRow` anatomy without the `IconWell`, a violet check in place of the
value". That is what `ListRow` already is: `trailing="check"` with `selected` draws the violet check
(`palette.violet.500`), and the row has never had an icon well. Building `SelectRow` would have
shipped a second name for a component that already exists, so it was not built. The `SelectRow`
rows in `SCREEN-STATES.md` (A5, D5) are `ListRow trailing="check"`.

```tsx
// The Currency list (A5, D5).
<DataRowList>
  <ListRow label="GBP · British Pound" trailing="check" selected />
  <ListRow label="IDR · Indonesian Rupiah" trailing="check" />
  <ListRow label="USD · US Dollar" trailing="check" />
</DataRowList>
```
