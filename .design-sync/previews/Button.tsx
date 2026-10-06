import type { ReactNode } from 'react';
import { Button } from '@astrisk/ds';

// Astrisk is dark-only: show every story on the app's page surface.
const Night = ({ children }: { children: ReactNode }) => (
  <div style={{ background: 'var(--c-bg)', color: 'var(--c-text)', fontFamily: 'var(--font-body)', padding: 24, borderRadius: 12, display: 'flow-root' }}>{children}</div>
);

const row = { display: 'flex', flexWrap: 'wrap' as const, gap: 12, alignItems: 'center' };

const VariantsStory = () => (
  <div style={row}>
    <Button variant="primary">Create kundali</Button>
    <Button variant="secondary">Save chart</Button>
    <Button variant="ghost">Cancel</Button>
    <Button variant="danger">Delete account</Button>
    <Button variant="link">View full reading</Button>
  </div>
);
const SmallStory = () => (
  <div style={row}>
    <Button variant="primary" size="sm">Send interest</Button>
    <Button size="sm">Shortlist</Button>
    <Button variant="ghost" size="sm">Not now</Button>
  </div>
);
const StatesStory = () => (
  <div style={row}>
    <Button variant="primary" busy>Calculating</Button>
    <Button disabled>Unavailable</Button>
    <Button icon aria-label="Next month">›</Button>
  </div>
);
const BlockStory = () => <div style={{ width: 320 }}><Button variant="primary" block>Continue</Button></div>;

export const Variants = () => <Night><VariantsStory /></Night>;
export const Small = () => <Night><SmallStory /></Night>;
export const States = () => <Night><StatesStory /></Night>;
export const Block = () => <Night><BlockStory /></Night>;
