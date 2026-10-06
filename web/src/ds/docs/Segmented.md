---
category: Navigation
---

# Segmented

A compact group of mutually exclusive view options (aria-pressed buttons).

## Usage

For switching how the same content is shown: chart style, reading method, units. Not for page navigation (use Tabs).

```tsx
<Segmented label="Chart style" value={v} onChange={setV} options={[['south','South'],['north','North']]} />
```
