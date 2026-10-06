import type { ReactNode } from 'react';
import { Avatar } from '@astrisk/ds';

// Astrisk is dark-only: show every story on the app's page surface.
const Night = ({ children }: { children: ReactNode }) => (
  <div style={{ background: 'var(--c-bg)', color: 'var(--c-text)', fontFamily: 'var(--font-body)', padding: 24, borderRadius: 12, display: 'flow-root' }}>{children}</div>
);

const SizesStory = () => (
  <div style={{ display: 'flex', alignItems: 'center', gap: 14 }}>
    <Avatar name="Ananya Rao" size={28} /><Avatar name="Kiran Shetty" /><Avatar name="@meera" size={48} /><Avatar name="Vikram Iyer" size={64} />
  </div>
);
const InListStory = () => (
  <div style={{ display: 'grid', gap: 10 }}>
    {['Ananya Rao', 'Kiran Shetty', 'Meera Nair'].map(n => (
      <div key={n} style={{ display: 'flex', alignItems: 'center', gap: 10 }}><Avatar name={n} /><span>{n}</span></div>
    ))}
  </div>
);

export const Sizes = () => <Night><SizesStory /></Night>;
export const InList = () => <Night><InListStory /></Night>;
