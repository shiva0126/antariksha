import type { ReactNode } from 'react';
import { Chip } from '@astrisk/ds';

// Astrisk is dark-only: show every story on the app's page surface.
const Night = ({ children }: { children: ReactNode }) => (
  <div style={{ background: 'var(--c-bg)', color: 'var(--c-text)', fontFamily: 'var(--font-body)', padding: 24, borderRadius: 12, display: 'flow-root' }}>{children}</div>
);

const TonesStory = () => (
  <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8 }}>
    <Chip>Neutral</Chip><Chip tone="accent">Verified</Chip><Chip tone="success">Accepted</Chip>
    <Chip tone="warning">Pending</Chip><Chip tone="danger">Mangal dosha</Chip><Chip tone="info">Shared horoscope</Chip>
  </div>
);
const ClickableStory = () => (
  <div style={{ display: 'flex', gap: 8 }}>
    <Chip onClick={() => {}} title="Filter by Moon sign">Vrishabha ✕</Chip>
    <Chip tone="accent" onClick={() => {}}>Age 26–32 ✕</Chip>
  </div>
);

export const Tones = () => <Night><TonesStory /></Night>;
export const Clickable = () => <Night><ClickableStory /></Night>;
