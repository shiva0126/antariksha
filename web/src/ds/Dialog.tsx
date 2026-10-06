import { useEffect, useRef, type ReactNode } from 'react';

/** A modal dialog on the native <dialog> element: focus is trapped, Escape
 *  closes, and the page behind cannot be scrolled or clicked. */
export function Dialog({ open, onClose, title, children, wide }: { open: boolean; onClose: () => void; title: string; children: ReactNode; wide?: boolean }) {
  const ref = useRef<HTMLDialogElement>(null);
  useEffect(() => {
    const d = ref.current;
    if (!d) return;
    if (open && !d.open) d.showModal();
    if (!open && d.open) d.close();
  }, [open]);
  return (
    <dialog ref={ref} className={'dialog' + (wide ? ' dialog--wide' : '')} aria-label={title} onClose={onClose} onCancel={e => { e.preventDefault(); onClose(); }}
      onClick={e => { if (e.target === ref.current) onClose(); }}>
      <div className="dialog-body">
        <header className="dialog-head"><h2>{title}</h2><button type="button" className="ds-btn ds-btn--ghost ds-btn--sm" aria-label="Close" onClick={onClose}>✕</button></header>
        {open && children}
      </div>
    </dialog>
  );
}
