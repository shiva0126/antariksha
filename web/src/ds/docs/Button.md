---
category: Actions
---

# Button

The only button. Variants primary, secondary (default), ghost, danger, link; sizes md and sm.

## Usage

One primary per screen or card. danger only for destructive actions, and confirm them with useFeedback().confirm. `busy` shows a spinner and disables. `icon` for square icon-only buttons (give them aria-label). `block` for full width on phones and forms. Labels are verbs in sentence case ("Save chart", not "SAVE").

```tsx
<Button variant="primary" busy={saving}>Save chart</Button>
```
