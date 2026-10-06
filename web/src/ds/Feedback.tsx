import { createContext, useCallback, useContext, useRef, useState, type ReactNode } from 'react';
import { Button } from './Button';
import { Dialog } from './Dialog';
import { Field } from './Field';
import { Notice } from './Layout';

type Toast = { id: number; tone: 'success' | 'danger' | 'neutral'; text: string };
type Ask = { kind: 'confirm' | 'prompt'; title: string; body?: ReactNode; confirm: string; danger?: boolean; label?: string; initial?: string; max?: number; resolve: (v: unknown) => void };
type Ctx = {
  toast: (text: string, tone?: Toast['tone']) => void;
  confirm: (o: { title: string; body?: ReactNode; confirm?: string; danger?: boolean }) => Promise<boolean>;
  prompt: (o: { title: string; body?: ReactNode; label: string; confirm?: string; initial?: string; max?: number }) => Promise<string | null>;
};
const FeedbackCtx = createContext<Ctx | null>(null);

/** Toasts for results, and in-app confirm/prompt dialogs instead of the
 *  browser's own confirm and prompt boxes. Wrap the app once. */
export function FeedbackProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([]);
  const [ask, setAsk] = useState<Ask | null>(null);
  const [text, setText] = useState('');
  const next = useRef(0);
  const toast = useCallback((t: string, tone: Toast['tone'] = 'success') => {
    const id = ++next.current;
    setToasts(x => [...x.slice(-2), { id, tone, text: t }]);
    setTimeout(() => setToasts(x => x.filter(y => y.id !== id)), tone === 'danger' ? 7000 : 4000);
  }, []);
  const confirm: Ctx['confirm'] = useCallback(o => new Promise<boolean>(resolve => setAsk({ kind: 'confirm', title: o.title, body: o.body, confirm: o.confirm ?? 'Confirm', danger: o.danger, resolve: v => resolve(v as boolean) })), []);
  const prompt: Ctx['prompt'] = useCallback(o => new Promise<string | null>(resolve => { setText(o.initial ?? ''); setAsk({ kind: 'prompt', title: o.title, body: o.body, label: o.label, confirm: o.confirm ?? 'Save', max: o.max, resolve: v => resolve(v as string | null) }); }), []);
  const close = (v: unknown) => { ask?.resolve(v); setAsk(null); };
  return (
    <FeedbackCtx.Provider value={{ toast, confirm, prompt }}>
      {children}
      <div className="ds-toasts" aria-live="polite">{toasts.map(t => <div key={t.id} className="ds-toast"><Notice tone={t.tone}>{t.text}</Notice></div>)}</div>
      <Dialog open={!!ask} onClose={() => close(ask?.kind === 'confirm' ? false : null)} title={ask?.title ?? ''}>
        {ask && <form onSubmit={e => { e.preventDefault(); close(ask.kind === 'confirm' ? true : text.trim()); }} className="ds-stack">
          {ask.body && <div className="ds-muted">{ask.body}</div>}
          {ask.kind === 'prompt' && <Field label={ask.label!}><textarea autoFocus maxLength={ask.max ?? 500} value={text} onChange={e => setText(e.target.value)} rows={3} /></Field>}
          <div className="dialog-actions">
            <Button variant="ghost" onClick={() => close(ask.kind === 'confirm' ? false : null)}>Cancel</Button>
            <Button type="submit" variant={ask.danger ? 'danger' : 'primary'}>{ask.confirm}</Button>
          </div>
        </form>}
      </Dialog>
    </FeedbackCtx.Provider>
  );
}

export function useFeedback(): Ctx {
  const c = useContext(FeedbackCtx);
  if (!c) throw new Error('useFeedback outside FeedbackProvider');
  return c;
}
