import { useEffect, useState } from 'react';
import { KundaliPage } from './ui/kundali/KundaliPage';
import { PanchangPage } from './ui/panchang/PanchangPage';

type Page = 'kundali' | 'day' | 'month';
const pages: [Page, string][] = [['kundali', 'Kundali'], ['day', 'Daily Panchang'], ['month', 'Hindu Calendar']];
const fromHash = (): Page => { const p = location.hash.slice(1); return p === 'day' || p === 'month' ? p : 'kundali'; };

export default function App() {
  const [page, setPage] = useState<Page>(fromHash);
  const [menu, setMenu] = useState(false);
  useEffect(() => {
    const changed = () => setPage(fromHash());
    window.addEventListener('hashchange', changed);
    return () => window.removeEventListener('hashchange', changed);
  }, []);
  function go(p: Page) { location.hash = p; setPage(p); setMenu(false); window.scrollTo({ top: 0 }); }
  return (
    <div className="app">
      <header className="site-header">
        <button className="brand" onClick={() => go('kundali')} aria-label="Antariksha home"><span className="brand-mark" aria-hidden>अ</span><b>Antariksha</b></button>
        <button className="menu-toggle" aria-expanded={menu} aria-controls="main-nav" onClick={() => setMenu(!menu)}>Menu</button>
        <nav id="main-nav" className={'main-nav' + (menu ? ' open' : '')} aria-label="Main navigation">
          {pages.map(([id, label]) => <button key={id} className={page === id ? 'active' : ''} aria-current={page === id ? 'page' : undefined} onClick={() => go(id)}>{label}</button>)}
        </nav>
      </header>
      <main>{page === 'kundali' ? <KundaliPage /> : <PanchangPage key={page} mode={page} />}</main>
      <footer className="site-footer">Antariksha · Swiss Ephemeris · Lahiri ayanamsa · For reflection, not certainty.</footer>
    </div>
  );
}
