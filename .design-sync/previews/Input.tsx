import type { ReactNode } from 'react';
import { Input } from '@astrisk/ds';

// Astrisk is dark-only: show every story on the app's page surface.
const Night = ({ children }: { children: ReactNode }) => (
  <div style={{ background: 'var(--c-bg)', color: 'var(--c-text)', fontFamily: 'var(--font-body)', padding: 24, borderRadius: 12, display: 'flow-root' }}>{children}</div>
);

const TextStory = () => <div style={{ width: 320 }}><Input placeholder="Search people by name or handle" /></div>;
const DateAndTimeStory = () => (
  <div style={{ display: 'flex', gap: 12, width: 360 }}>
    <Input type="date" defaultValue="1994-08-17" aria-label="Birth date" />
    <Input type="time" defaultValue="06:42" aria-label="Birth time" />
  </div>
);
const DisabledStory = () => <div style={{ width: 320 }}><Input disabled defaultValue="member_ananya" aria-label="Handle" /></div>;

export const Text = () => <Night><TextStory /></Night>;
export const DateAndTime = () => <Night><DateAndTimeStory /></Night>;
export const Disabled = () => <Night><DisabledStory /></Night>;
