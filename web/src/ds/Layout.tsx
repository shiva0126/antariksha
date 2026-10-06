import type { ReactNode } from 'react';

export function PageHeader({ kicker, title, description, actions, compactOnMobile }: { kicker?: ReactNode; title: ReactNode; description?: ReactNode; actions?: ReactNode; compactOnMobile?: boolean }) {
  return (
    <header className={'ds-page-header' + (compactOnMobile ? ' ds-page-header--compact-mobile' : '')}>
      <div className="ds-page-header__text">
        {kicker && <p className="ds-kicker">{kicker}</p>}
        <h1>{title}</h1>
        {description && <p>{description}</p>}
      </div>
      {actions && <div className="ds-page-header__actions">{actions}</div>}
    </header>
  );
}

export function Card({ title, sub, actions, children, flush, tight, as: Tag = 'section', className, ...rest }: { title?: ReactNode; sub?: ReactNode; actions?: ReactNode; children?: ReactNode; flush?: boolean; tight?: boolean; as?: 'section' | 'article' | 'div' | 'aside'; className?: string; 'aria-label'?: string; id?: string }) {
  return (
    <Tag className={['ds-card', flush && 'ds-card--flush', tight && 'ds-card--tight', className].filter(Boolean).join(' ')} {...rest}>
      {(title || actions) && (
        <div className="ds-card__head">
          <div>{title && <h2 className="ds-card__title">{title}</h2>}{sub && <p className="ds-card__sub">{sub}</p>}</div>
          {actions && <div className="ds-row">{actions}</div>}
        </div>
      )}
      {children}
    </Tag>
  );
}

export function StatTile({ label, value, note }: { label: ReactNode; value: ReactNode; note?: ReactNode }) {
  return <div className="ds-stat"><span className="ds-stat__label">{label}</span><b className="ds-stat__value">{value}</b>{note && <span className="ds-stat__note">{note}</span>}</div>;
}

export type Tone = 'neutral' | 'accent' | 'success' | 'warning' | 'danger' | 'info';
export function Chip({ tone = 'neutral', children, onClick, title }: { tone?: Tone; children: ReactNode; onClick?: () => void; title?: string }) {
  const c = 'ds-chip' + (tone === 'neutral' ? '' : ` ds-chip--${tone}`);
  return onClick ? <button type="button" className={c} onClick={onClick} title={title}>{children}</button> : <span className={c} title={title}>{children}</span>;
}

export function EmptyState({ icon, title, children, action }: { icon?: ReactNode; title: ReactNode; children?: ReactNode; action?: ReactNode }) {
  return <div className="ds-empty">{icon}<h3>{title}</h3>{children && <p>{children}</p>}{action}</div>;
}

export function Notice({ tone = 'neutral', children }: { tone?: 'neutral' | 'success' | 'danger' | 'warning'; children: ReactNode }) {
  return <div className={'ds-notice' + (tone === 'neutral' ? '' : ` ds-notice--${tone}`)} role={tone === 'danger' ? 'alert' : 'status'}>{children}</div>;
}

export function Skeleton({ lines = 3 }: { lines?: number }) {
  return <div className="ds-skel" aria-label="Loading" role="status">{Array.from({ length: lines }, (_, i) => <i key={i} style={{ width: `${[70, 92, 56, 84, 64][i % 5]}%` }} />)}</div>;
}

export function Avatar({ name, size = 36, src }: { name: string; size?: number; src?: string }) {
  const initial = (name.replace(/^@/, '').trim()[0] || '?').toUpperCase();
  return <span className="ds-avatar" style={{ width: size, height: size, fontSize: size * 0.42 }} aria-hidden>{src ? <img src={src} alt="" /> : initial}</span>;
}

/** Prefer a person's display name; a handle only when there is no name. */
export const personName = (p: { display_name?: string | null; name?: string | null; handle?: string }) =>
  (p.display_name || p.name || '').trim() || (p.handle && !/^member_[0-9a-f]{12}$/.test(p.handle) ? '@' + p.handle : 'Member');
