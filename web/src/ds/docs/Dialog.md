---
category: Overlays
---

# Dialog

Modal dialog on the native <dialog>: focus trapped, Escape and backdrop click close it.

## Usage

Controlled with `open`/`onClose`. Put actions in a `<div className="dialog-actions">` at the end, cancel (ghost) before the main action. For simple confirm/prompt use useFeedback() instead of building a Dialog.

```tsx
<Dialog open={open} onClose={close} title="Delete this chart?">...</Dialog>
```
