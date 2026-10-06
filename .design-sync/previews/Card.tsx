import type { ReactNode } from 'react';
import { Button, Card, Chip } from '@astrisk/ds';

// Astrisk is dark-only: show every story on the app's page surface.
const Night = ({ children }: { children: ReactNode }) => (
  <div style={{ background: 'var(--c-bg)', color: 'var(--c-text)', fontFamily: 'var(--font-body)', padding: 24, borderRadius: 12, display: 'flow-root' }}>{children}</div>
);

const WithHeaderStory = () => (
  <div style={{ width: 440 }}>
    <Card title="Current dasha" sub="Vimshottari" actions={<Button size="sm" variant="ghost">Timeline</Button>}>
      <p>Jupiter mahadasha, Saturn antardasha, until 14 March 2028.</p>
    </Card>
  </div>
);
const TightStory = () => (
  <div style={{ width: 340 }}>
    <Card tight title="Mangal dosha"><div style={{ display: 'flex', gap: 8 }}><Chip tone="success">Not present</Chip><Chip>Lagna and Moon checked</Chip></div></Card>
  </div>
);
const PlainStory = () => (
  <div style={{ width: 440 }}>
    <Card><p>Rahu kalam today is 15:00 to 16:30. Avoid starting new work in this window.</p></Card>
  </div>
);

export const WithHeader = () => <Night><WithHeaderStory /></Night>;
export const Tight = () => <Night><TightStory /></Night>;
export const Plain = () => <Night><PlainStory /></Night>;
