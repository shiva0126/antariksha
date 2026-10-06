import type { ReactNode } from 'react';
import { Card, Skeleton } from '@astrisk/ds';

// Astrisk is dark-only: show every story on the app's page surface.
const Night = ({ children }: { children: ReactNode }) => (
  <div style={{ background: 'var(--c-bg)', color: 'var(--c-text)', fontFamily: 'var(--font-body)', padding: 24, borderRadius: 12, display: 'flow-root' }}>{children}</div>
);

const LoadingStory = () => <div style={{ width: 420 }}><Card title="Reading"><Skeleton lines={4} /></Card></div>;
const ShortStory = () => <div style={{ width: 300 }}><Skeleton lines={2} /></div>;

export const Loading = () => <Night><LoadingStory /></Night>;
export const Short = () => <Night><ShortStory /></Night>;
