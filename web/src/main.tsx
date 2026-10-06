import React from 'react';
import { createRoot } from 'react-dom/client';
// Order matters: tokens, legacy page styles, then the design system on top.
import './ds/tokens.css';
import './styles.css';
import './ds/ds.css';
import App from './App';

createRoot(document.getElementById('root')!).render(<React.StrictMode><App /></React.StrictMode>);

// The service worker enables "Add to Home Screen" and push alerts.
if ('serviceWorker' in navigator && location.protocol === 'https:') {
  window.addEventListener('load', () => { navigator.serviceWorker.register('/sw.js').catch(() => {}); });
}
