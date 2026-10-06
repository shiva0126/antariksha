import type { ReactNode } from 'react';
import { Button, PageHeader } from '@astrisk/ds';

// Astrisk is dark-only: show every story on the app's page surface.
const Night = ({ children }: { children: ReactNode }) => (
  <div style={{ background: 'var(--c-bg)', color: 'var(--c-text)', fontFamily: 'var(--font-body)', padding: 24, borderRadius: 12, display: 'flow-root' }}>{children}</div>
);

const WithActionsStory = () => (
  <div style={{ width: 760 }}>
    <PageHeader kicker="Kundali" title="Ananya Rao" description="Born 17 August 1994, 06:42 in Udupi, Karnataka"
      actions={<><Button>Edit birth details</Button><Button variant="primary">Save chart</Button></>} />
  </div>
);
const TitleOnlyStory = () => (
  <div style={{ width: 760 }}>
    <PageHeader title="Today's panchang" description="Tuesday, 6 October 2026 · Bengaluru" />
  </div>
);

export const WithActions = () => <Night><WithActionsStory /></Night>;
export const TitleOnly = () => <Night><TitleOnlyStory /></Night>;
