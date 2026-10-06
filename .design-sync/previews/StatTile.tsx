import type { ReactNode } from 'react';
import { StatTile } from '@astrisk/ds';

// Astrisk is dark-only: show every story on the app's page surface.
const Night = ({ children }: { children: ReactNode }) => (
  <div style={{ background: 'var(--c-bg)', color: 'var(--c-text)', fontFamily: 'var(--font-body)', padding: 24, borderRadius: 12, display: 'flow-root' }}>{children}</div>
);

const RowStory = () => (
  <div style={{ display: 'flex', gap: 12 }}>
    <StatTile label="Guna milan" value="27 / 36" note="Good match" />
    <StatTile label="Moon sign" value="Vrishabha" note="Rohini nakshatra" />
    <StatTile label="Life path" value="7" />
  </div>
);
const SingleStory = () => <div style={{ width: 200 }}><StatTile label="Sunrise" value="06:09" note="Bengaluru" /></div>;

export const Row = () => <Night><RowStory /></Night>;
export const Single = () => <Night><SingleStory /></Night>;
