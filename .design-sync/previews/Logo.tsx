import type { ReactNode } from 'react';
import { Logo } from '@astrisk/ds';

// Astrisk is dark-only: show every story on the app's page surface.
const Night = ({ children }: { children: ReactNode }) => (
  <div style={{ background: 'var(--c-bg)', color: 'var(--c-text)', fontFamily: 'var(--font-body)', padding: 24, borderRadius: 12, display: 'flow-root' }}>{children}</div>
);

const HeaderStory = () => <Logo href="#kundali" />;
const CompactStory = () => <Logo href="#kundali" compact />;

export const Header = () => <Night><HeaderStory /></Night>;
export const Compact = () => <Night><CompactStory /></Night>;
