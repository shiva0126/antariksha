import type { ReactNode } from 'react';
import { ButtonLink } from '@astrisk/ds';

// Astrisk is dark-only: show every story on the app's page surface.
const Night = ({ children }: { children: ReactNode }) => (
  <div style={{ background: 'var(--c-bg)', color: 'var(--c-text)', fontFamily: 'var(--font-body)', padding: 24, borderRadius: 12, display: 'flow-root' }}>{children}</div>
);

const LinksStory = () => (
  <div style={{ display: 'flex', gap: 12 }}>
    <ButtonLink variant="primary" href="#matching">Check compatibility</ButtonLink>
    <ButtonLink href="#panchang/today">Today's panchang</ButtonLink>
    <ButtonLink variant="link" href="#me/privacy">Privacy and data</ButtonLink>
  </div>
);

export const Links = () => <Night><LinksStory /></Night>;
