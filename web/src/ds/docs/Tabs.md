---
category: Navigation
---

# Tabs

The one tab style. Items are `{id, label, count?}`; with `base` the tabs are links to `#base/id`.

## Usage

Use `base` for page sections so the back button and shared links work. `count` shows a small badge for new items. Always pass an accessible `label`.

```tsx
<Tabs label="Kundali" base="kundali" active={tab} items={items} />
```
