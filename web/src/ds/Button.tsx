import type { AnchorHTMLAttributes, ButtonHTMLAttributes, ReactNode } from 'react';

type Variant = 'primary' | 'secondary' | 'ghost' | 'danger' | 'link';
type Common = { variant?: Variant; size?: 'md' | 'sm'; block?: boolean; busy?: boolean; icon?: boolean; children?: ReactNode };

const cls = ({ variant = 'secondary', size = 'md', block, icon }: Common, extra?: string) =>
  ['ds-btn', `ds-btn--${variant}`, size === 'sm' && 'ds-btn--sm', block && 'ds-btn--block', icon && 'ds-btn--icon', extra].filter(Boolean).join(' ');

/** The only button. One primary per screen; `busy` shows a spinner and disables it. */
export function Button({ variant, size, block, busy, icon, children, className, disabled, type = 'button', ...rest }: Common & ButtonHTMLAttributes<HTMLButtonElement>) {
  return (
    <button type={type} className={cls({ variant, size, block, icon }, className)} disabled={disabled || busy} aria-busy={busy || undefined} {...rest}>
      {busy && <span className="ds-spinner" aria-hidden />}{children}
    </button>
  );
}

/** A link styled as a button (navigation, downloads). */
export function ButtonLink({ variant, size, block, children, className, ...rest }: Common & AnchorHTMLAttributes<HTMLAnchorElement>) {
  return <a className={cls({ variant, size, block }, className)} {...rest}>{children}</a>;
}
