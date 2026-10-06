import type { ReactNode } from 'react';
import { Notice } from '@astrisk/ds';

// Astrisk is dark-only: show every story on the app's page surface.
const Night = ({ children }: { children: ReactNode }) => (
  <div style={{ background: 'var(--c-bg)', color: 'var(--c-text)', fontFamily: 'var(--font-body)', padding: 24, borderRadius: 12, display: 'flow-root' }}>{children}</div>
);

const TonesStory = () => (
  <div style={{ display: 'grid', gap: 10, width: 460 }}>
    <Notice>Birth time is unknown, so the lagna and houses are approximate.</Notice>
    <Notice tone="success">Private profile saved.</Notice>
    <Notice tone="warning">Your photo is waiting for verification.</Notice>
    <Notice tone="danger">Could not reach the server. Check your connection and try again.</Notice>
  </div>
);

export const Tones = () => <Night><TonesStory /></Night>;
