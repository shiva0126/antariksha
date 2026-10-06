import type { ReactNode } from 'react';
import { Button, EmptyState, NorthStar } from '@astrisk/ds';

// Astrisk is dark-only: show every story on the app's page surface.
const Night = ({ children }: { children: ReactNode }) => (
  <div style={{ background: 'var(--c-bg)', color: 'var(--c-text)', fontFamily: 'var(--font-body)', padding: 24, borderRadius: 12, display: 'flow-root' }}>{children}</div>
);

const NoChartsStory = () => (
  <div style={{ width: 440 }}>
    <EmptyState icon={<NorthStar size={40} />} title="No saved charts yet" action={<Button variant="primary">Create a kundali</Button>}>
      Charts you save appear here so you can open them on any device.
    </EmptyState>
  </div>
);
const NoInterestsStory = () => (
  <div style={{ width: 440 }}>
    <EmptyState title="No interests yet">When someone sends you an interest, it shows up here.</EmptyState>
  </div>
);

export const NoCharts = () => <Night><NoChartsStory /></Night>;
export const NoInterests = () => <Night><NoInterestsStory /></Night>;
