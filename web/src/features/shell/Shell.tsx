import { useEffect, useRef, useState, type ReactNode } from 'react';
import { Avatar, Logo, NorthStar, personName } from '../../ds';
import { useT } from '../../i18n';
import { href, type Route, type Section } from '../../lib/router';
import { api } from '../../lib/api';
import './shell.css';

export type Me = { id: string; handle: string; email: string; role: string; profile?: { name?: string } };

const PRIMARY: [Section, string, ReactNode][] = [
  ['kundali', 'Kundali', <NorthStar key="k" size={20} ring={false} color="currentColor" />],
  ['panchang', 'Panchang', <svg key="p" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" aria-hidden><path d="M20 14.5A8 8 0 1 1 9.5 4a6.5 6.5 0 0 0 10.5 10.5z" /></svg>],
  ['matching', 'Matching', <svg key="m" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" aria-hidden><circle cx="9" cy="12" r="5" /><circle cx="15" cy="12" r="5" /></svg>],
  ['matrimony', 'Matrimony', <svg key="t" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinejoin="round" aria-hidden><path d="M12 20s-7-4.4-7-10a4 4 0 0 1 7-2.6A4 4 0 0 1 19 10c0 5.6-7 10-7 10z" /></svg>],
  ['community', 'Community', <svg key="c" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" aria-hidden><circle cx="9" cy="9" r="3" /><circle cx="17" cy="10" r="2.5" /><path d="M3 19c0-3 3-5 6-5s6 2 6 5" /><path d="M15 15c2.5 0 5 1.5 5 4" /></svg>],
];

/** Unread notifications, refreshed every minute and when the window regains focus. */
function useUnread() {
  const [n, setN] = useState(0);
  useEffect(() => {
    let live = true;
    const load = () => api<{ read: boolean }[]>('/api/me/notifications').then(r => { if (live) setN(r.filter(x => !x.read).length); }).catch(() => {});
    load();
    const t = setInterval(load, 60_000);
    window.addEventListener('focus', load);
    window.addEventListener('astrisk-notifications', load);
    return () => { live = false; clearInterval(t); window.removeEventListener('focus', load); window.removeEventListener('astrisk-notifications', load); };
  }, []);
  return n;
}

function MeMenu({ me, onSignOut }: { me: Me; onSignOut: () => void }) {
  const t = useT();
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);
  useEffect(() => {
    const close = (e: MouseEvent) => { if (!ref.current?.contains(e.target as Node)) setOpen(false); };
    const esc = (e: KeyboardEvent) => { if (e.key === 'Escape') setOpen(false); };
    document.addEventListener('mousedown', close); document.addEventListener('keydown', esc);
    window.addEventListener('hashchange', () => setOpen(false));
    return () => { document.removeEventListener('mousedown', close); document.removeEventListener('keydown', esc); };
  }, []);
  const name = personName({ name: me.profile?.name, handle: me.handle });
  const items: [string, string][] = [['profile', t('Profile')], ['charts', t('Saved charts')], ['notifications', t('Notifications')], ['security', t('Security')], ['privacy', t('Privacy and data')], ['settings', t('Settings')]];
  return (
    <div className="shell-me" ref={ref}>
      <button type="button" className="shell-me__toggle" aria-haspopup="menu" aria-expanded={open} aria-label={`${t('Me')}: ${name}`} onClick={() => setOpen(!open)}>
        <Avatar name={name} size={32} /><span className="shell-me__name">{name}</span>
      </button>
      {open && (
        <div className="shell-me__menu" role="menu">
          <div className="shell-me__who"><Avatar name={name} size={40} /><div><b>{name}</b><span>{me.email || '@' + me.handle}</span></div></div>
          {items.map(([sub, label]) => <a key={sub} role="menuitem" href={href('me', sub)}>{label}</a>)}
          {(me.role === 'superadmin' || me.role === 'moderator') && <a role="menuitem" href={href('admin')}>{t('Admin')}</a>}
          <button type="button" role="menuitem" onClick={onSignOut}>{t('Sign out')}</button>
        </div>
      )}
    </div>
  );
}

export function Shell({ route, me, onSignOut, children }: { route: Route; me: Me; onSignOut: () => void; children: ReactNode }) {
  const t = useT();
  const unread = useUnread();
  const current = route.section === 'biodata' ? 'matrimony' : route.section;
  return (
    <div className="shell">
      <a className="shell-skip" href="#main">Skip to content</a>
      <header className="shell-header">
        <div className="shell-header__inner">
          <Logo />
          <nav className="shell-nav" aria-label={t('Sections')}>
            {PRIMARY.map(([id, label]) => <a key={id} href={href(id)} aria-current={current === id ? 'page' : undefined}>{t(label)}</a>)}
          </nav>
          <div className="shell-tools">
            <a className="shell-bell" href={href('me', 'notifications')} aria-label={unread ? `${t('Notifications')}, ${unread} new` : t('Notifications')}>
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" aria-hidden><path d="M6 8a6 6 0 1 1 12 0c0 7 3 9 3 9H3s3-2 3-9" /><path d="M10.3 21a1.94 1.94 0 0 0 3.4 0" /></svg>
              {unread > 0 && <span className="ds-count">{unread > 9 ? '9+' : unread}</span>}
            </a>
            <MeMenu me={me} onSignOut={onSignOut} />
          </div>
        </div>
      </header>
      <main id="main" className="shell-main">{children}</main>
      <footer className="shell-footer">
        <NorthStar size={16} ring={false} color="var(--c-accent-line)" />
        <span>Astrisk is free: no payments, no remedies for sale.</span>
        <a href={href('me', 'privacy')}>{t('Privacy and data')}</a>
        <span>Swiss Ephemeris · GeoNames (CC BY 4.0) · For reflection, not certainty.</span>
      </footer>
      <nav className="shell-tabbar" aria-label={t('Sections')}>
        {PRIMARY.map(([id, label, icon]) => <a key={id} href={href(id)} aria-current={current === id ? 'page' : undefined}>{icon}<span>{t(label)}</span></a>)}
      </nav>
    </div>
  );
}
