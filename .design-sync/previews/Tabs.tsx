import type { ReactNode } from 'react';
import { Tabs } from '@astrisk/ds';

// Astrisk is dark-only: show every story on the app's page surface.
const Night = ({ children }: { children: ReactNode }) => (
  <div style={{ background: 'var(--c-bg)', color: 'var(--c-text)', fontFamily: 'var(--font-body)', padding: 24, borderRadius: 12, display: 'flow-root' }}>{children}</div>
);

const KundaliTabsStory = () => (
  <Tabs label="Kundali" active="chart" items={[
    { id: 'overview', label: 'Overview' }, { id: 'chart', label: 'Chart' }, { id: 'planets', label: 'Planets' },
    { id: 'dasha', label: 'Dasha' }, { id: 'reading', label: 'Reading' }, { id: 'ask', label: 'Ask Astrisk' }, { id: 'systems', label: 'Other systems' },
  ]} />
);
const WithCountsStory = () => (
  <Tabs label="Matrimony" base="matrimony" active="discover" items={[
    { id: 'discover', label: 'Discover' }, { id: 'interests', label: 'Interests', count: 3 }, { id: 'chats', label: 'Chats', count: 1 }, { id: 'profile', label: 'My profile' },
  ]} />
);

export const KundaliTabs = () => <Night><KundaliTabsStory /></Night>;
export const WithCounts = () => <Night><WithCountsStory /></Night>;
