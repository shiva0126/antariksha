import type { ReactNode } from 'react';
import { Checkbox } from '@astrisk/ds';

// Astrisk is dark-only: show every story on the app's page surface.
const Night = ({ children }: { children: ReactNode }) => (
  <div style={{ background: 'var(--c-bg)', color: 'var(--c-text)', fontFamily: 'var(--font-body)', padding: 24, borderRadius: 12, display: 'flow-root' }}>{children}</div>
);

const OptionsStory = () => (
  <div style={{ display: 'grid', gap: 10 }}>
    <Checkbox label="Show my horoscope match to people I accept" defaultChecked />
    <Checkbox label="Email me about new interests" />
    <Checkbox label="Hide my profile from search" disabled />
  </div>
);

export const Options = () => <Night><OptionsStory /></Night>;
