---
category: Forms
---

# Field

Label above, control, then hint or error below. Wraps exactly one control and wires id, aria-invalid and aria-describedby.

## Usage

Every input gets a Field. Use `error` for validation messages (replaces the hint), `optional` instead of asterisks, `span` to take both columns in a FormGrid.

```tsx
<Field label="Birth time" hint="24-hour clock"><Input type="time" /></Field>
```
