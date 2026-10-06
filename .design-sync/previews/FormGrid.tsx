import type { ReactNode } from 'react';
import { Field, FormGrid, Input, Select } from '@astrisk/ds';

// Astrisk is dark-only: show every story on the app's page surface.
const Night = ({ children }: { children: ReactNode }) => (
  <div style={{ background: 'var(--c-bg)', color: 'var(--c-text)', fontFamily: 'var(--font-body)', padding: 24, borderRadius: 12, display: 'flow-root' }}>{children}</div>
);

const BirthDetailsStory = () => (
  <div style={{ width: 560 }}>
    <FormGrid>
      <Field label="Name"><Input defaultValue="Ananya Rao" /></Field>
      <Field label="Gender"><Select defaultValue="f" options={[['f', 'Female'], ['m', 'Male']]} /></Field>
      <Field label="Birth date"><Input type="date" defaultValue="1994-08-17" /></Field>
      <Field label="Birth time"><Input type="time" defaultValue="06:42" /></Field>
      <Field label="Birthplace" span hint="Pick from the list so the chart uses the right time zone."><Input defaultValue="Udupi, Karnataka, India" /></Field>
    </FormGrid>
  </div>
);

export const BirthDetails = () => <Night><BirthDetailsStory /></Night>;
