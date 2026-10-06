import type { ReactNode } from 'react';
import { Button, FeedbackProvider, useFeedback } from '@astrisk/ds';

// Astrisk is dark-only: show every story on the app's page surface.
const Night = ({ children }: { children: ReactNode }) => (
  <div style={{ background: 'var(--c-bg)', color: 'var(--c-text)', fontFamily: 'var(--font-body)', padding: 24, borderRadius: 12, display: 'flow-root' }}>{children}</div>
);

function Actions() {
  const fb = useFeedback();
  return (
    <div style={{ display: 'flex', gap: 12 }}>
      <Button variant="primary" onClick={() => fb.toast('Chart saved.')}>Save chart</Button>
      <Button variant="danger" onClick={() => fb.confirm({ title: 'Delete account?', body: 'Your charts and profile will be removed.', confirm: 'Delete', danger: true })}>Delete account</Button>
      <Button onClick={() => fb.prompt({ title: 'Suggest this profile', label: 'Note for your family' })}>Suggest</Button>
    </div>
  );
}

const ToastAndDialogsStory = () => <FeedbackProvider><Actions /></FeedbackProvider>;

export const ToastAndDialogs = () => <Night><ToastAndDialogsStory /></Night>;
