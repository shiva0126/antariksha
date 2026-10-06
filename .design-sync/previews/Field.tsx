import type { ReactNode } from 'react';
import { Field, Input, Select, Textarea } from '@astrisk/ds';

// Astrisk is dark-only: show every story on the app's page surface.
const Night = ({ children }: { children: ReactNode }) => (
  <div style={{ background: 'var(--c-bg)', color: 'var(--c-text)', fontFamily: 'var(--font-body)', padding: 24, borderRadius: 12, display: 'flow-root' }}>{children}</div>
);

const w = { width: 360, display: 'grid', gap: 16 };

const TextFieldStory = () => (
  <div style={w}>
    <Field label="Full name" hint="Shown on your kundali report."><Input defaultValue="Ananya Rao" /></Field>
  </div>
);
const WithErrorStory = () => (
  <div style={w}>
    <Field label="Birth time" error="Enter a time like 06:42."><Input defaultValue="6.42" /></Field>
  </div>
);
const OptionalStory = () => (
  <div style={w}>
    <Field label="Gotra" optional><Input placeholder="e.g. Kashyapa" /></Field>
    <Field label="About me" optional hint="Up to 300 characters."><Textarea defaultValue="Software engineer in Bengaluru. I enjoy Carnatic music and long treks." /></Field>
  </div>
);
const WithSelectStory = () => (
  <div style={w}>
    <Field label="Month system"><Select defaultValue="amanta" options={[['amanta', 'Amanta (new moon)'], ['purnimanta', 'Purnimanta (full moon)']]} /></Field>
  </div>
);

export const TextField = () => <Night><TextFieldStory /></Night>;
export const WithError = () => <Night><WithErrorStory /></Night>;
export const Optional = () => <Night><OptionalStory /></Night>;
export const WithSelect = () => <Night><WithSelectStory /></Night>;
