import type { ReactNode } from 'react';
import { Select } from '@astrisk/ds';

// Astrisk is dark-only: show every story on the app's page surface.
const Night = ({ children }: { children: ReactNode }) => (
  <div style={{ background: 'var(--c-bg)', color: 'var(--c-text)', fontFamily: 'var(--font-body)', padding: 24, borderRadius: 12, display: 'flow-root' }}>{children}</div>
);

const LanguageStory = () => (
  <div style={{ width: 260 }}>
    <Select aria-label="Language" defaultValue="en" options={[['en', 'English'], ['hi', 'हिन्दी'], ['kn', 'ಕನ್ನಡ'], ['ta', 'தமிழ்'], ['te', 'తెలుగు']]} />
  </div>
);
const WithPlaceholderStory = () => (
  <div style={{ width: 260 }}>
    <Select aria-label="Rashi" defaultValue="" placeholder="Any Moon sign" options={[['mesha', 'Mesha'], ['vrishabha', 'Vrishabha'], ['mithuna', 'Mithuna']]} />
  </div>
);

export const Language = () => <Night><LanguageStory /></Night>;
export const WithPlaceholder = () => <Night><WithPlaceholderStory /></Night>;
