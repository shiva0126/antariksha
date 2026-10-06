import type { ReactNode } from 'react';

export type TabItem<K extends string> = { id: K; label: ReactNode; count?: number };

/** One tab style everywhere. Tabs are links (`#section/tab`) so the back
 *  button and shared links work; `onSelect` is called as well. */
export function Tabs<K extends string>({ items, active, base, onSelect, label }: { items: TabItem<K>[]; active: K; base?: string; onSelect?: (id: K) => void; label: string }) {
  return (
    <nav className="ds-tabs" aria-label={label} role="tablist">
      {items.map(t => base ? (
        <a key={t.id} role="tab" className="ds-tabs__tab" aria-selected={t.id === active} href={`#${base}/${t.id}`} onClick={() => onSelect?.(t.id)}>
          {t.label}{!!t.count && <span className="ds-count" aria-label={`${t.count} new`}>{t.count}</span>}
        </a>
      ) : (
        <button key={t.id} type="button" role="tab" className="ds-tabs__tab" aria-selected={t.id === active} onClick={() => onSelect?.(t.id)}>
          {t.label}{!!t.count && <span className="ds-count" aria-label={`${t.count} new`}>{t.count}</span>}
        </button>
      ))}
    </nav>
  );
}

/** A small set of mutually exclusive view options (chart style, units). */
export function Segmented<K extends string>({ options, value, onChange, label }: { options: [K, ReactNode][]; value: K; onChange: (v: K) => void; label: string }) {
  return (
    <div className="ds-seg" role="group" aria-label={label}>
      {options.map(([k, l]) => <button key={k} type="button" aria-pressed={k === value} onClick={() => onChange(k)}>{l}</button>)}
    </div>
  );
}
