import type { ReactNode } from 'react';
import { Textarea } from '@astrisk/ds';

// Astrisk is dark-only: show every story on the app's page surface.
const Night = ({ children }: { children: ReactNode }) => (
  <div style={{ background: 'var(--c-bg)', color: 'var(--c-text)', fontFamily: 'var(--font-body)', padding: 24, borderRadius: 12, display: 'flow-root' }}>{children}</div>
);

const NoteStory = () => (
  <div style={{ width: 380 }}>
    <Textarea aria-label="Note with your interest" defaultValue="Namaste! Our families are both from Udupi and I liked your profile. Would you like to talk?" />
  </div>
);
const EmptyStory = () => <div style={{ width: 380 }}><Textarea aria-label="Ask Astrisk" placeholder="Ask about your chart, dasha or today's muhurta" /></div>;

export const Note = () => <Night><NoteStory /></Night>;
export const Empty = () => <Night><EmptyStory /></Night>;
