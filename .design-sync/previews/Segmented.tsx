import { useState, type ReactNode } from 'react';
import { Segmented } from '@astrisk/ds';

// Astrisk is dark-only: show every story on the app's page surface.
const Night = ({ children }: { children: ReactNode }) => (
  <div style={{ background: 'var(--c-bg)', color: 'var(--c-text)', fontFamily: 'var(--font-body)', padding: 24, borderRadius: 12, display: 'flow-root' }}>{children}</div>
);

const ChartStyleStory = () => {
  const [v, setV] = useState<'south' | 'north' | 'east'>('south');
  return <Segmented label="Chart style" value={v} onChange={setV} options={[['south', 'South Indian'], ['north', 'North Indian'], ['east', 'East Indian']]} />;
};
const MethodStory = () => {
  const [v, setV] = useState<'vedic' | 'numerology'>('numerology');
  return <Segmented label="Reading method" value={v} onChange={setV} options={[['vedic', 'Vedic'], ['numerology', 'Numerology']]} />;
};

export const ChartStyle = () => <Night><ChartStyleStory /></Night>;
export const Method = () => <Night><MethodStory /></Night>;
