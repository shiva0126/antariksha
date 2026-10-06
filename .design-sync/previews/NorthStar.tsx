import type { ReactNode } from 'react';
import { NorthStar } from '@astrisk/ds';

// Astrisk is dark-only: show every story on the app's page surface.
const Night = ({ children }: { children: ReactNode }) => (
  <div style={{ background: 'var(--c-bg)', color: 'var(--c-text)', fontFamily: 'var(--font-body)', padding: 24, borderRadius: 12, display: 'flow-root' }}>{children}</div>
);

const WithRingStory = () => <NorthStar size={96} title="Astrisk north star" />;
const StarOnlyStory = () => <NorthStar size={96} ring={false} title="North star without ring" />;
const SizesStory = () => (
  <div style={{ display: 'flex', alignItems: 'center', gap: 20 }}>
    <NorthStar size={20} /><NorthStar size={32} /><NorthStar size={56} /><NorthStar size={88} />
  </div>
);
const OnLightStory = () => (
  <div style={{ background: '#f4ead2', padding: 20, borderRadius: 12, display: 'inline-flex' }}>
    <NorthStar size={72} color="#0d1322" ringColor="#0d132255" title="North star in ink" />
  </div>
);

export const WithRing = () => <Night><WithRingStory /></Night>;
export const StarOnly = () => <Night><StarOnlyStory /></Night>;
export const Sizes = () => <Night><SizesStory /></Night>;
export const OnLight = () => <Night><OnLightStory /></Night>;
