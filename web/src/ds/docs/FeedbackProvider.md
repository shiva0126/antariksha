---
category: Overlays
---

# FeedbackProvider

Provides useFeedback(): toast(text, tone), confirm({title, body, confirm, danger}) and prompt({title, label}) as in-app dialogs.

## Usage

Wrap the app once. Never use window.alert/confirm/prompt. confirm resolves to a boolean, prompt to the text or null.

```tsx
const fb = useFeedback(); if (await fb.confirm({ title: "Delete?", danger: true })) remove();
```
