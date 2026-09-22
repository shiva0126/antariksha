import { useEffect, useState } from 'react';
import { SettingsProvider, useSettings, useT } from './i18n';
import { KundaliPage } from './ui/kundali/KundaliPage';
import { MatchPage } from './ui/match/MatchPage';
import { MuhurtaPage } from './ui/muhurta/MuhurtaPage';
import { PanchangPage } from './ui/panchang/PanchangPage';
import { PrivacyPage } from './ui/privacy/PrivacyPage';

type Page = 'kundali' | 'match' | 'muhurta' | 'day' | 'month' | 'privacy';
const pages: [Page, string][] = [['kundali', 'Kundali'], ['match', 'Matching'], ['muhurta', 'Muhurta'], ['day', 'Daily Panchang'], ['month', 'Hindu Calendar']];
const valid = new Set<string>(['kundali', 'match', 'muhurta', 'day', 'month', 'privacy']);
const fromHash = (): Page => { const p = location.hash.slice(1); return (valid.has(p) ? p : 'kundali') as Page; };

function Shell() {
  const t = useT();
  const { lang, setLang, months, setMonths } = useSettings();
  const [page, setPage] = useState<Page>(fromHash);
  const [menu, setMenu] = useState(false);
  useEffect(() => {
    const changed = () => { setPage(fromHash()); window.scrollTo({ top: 0 }); };
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
          {pages.map(([id, label]) => <button key={id} className={page === id ? 'active' : ''} aria-current={page === id ? 'page' : undefined} onClick={() => go(id)}>{t(label)}</button>)}
          <div className="settings">
            <label><span className="sr-only">{t('Language')}</span>
              <select aria-label="Language" value={lang} onChange={e => setLang(e.target.value as 'en' | 'hi')}><option value="en">English</option><option value="hi">हिन्दी</option></select></label>
            <label><span className="sr-only">{t('Months')}</span>
              <select aria-label="Month system" value={months} onChange={e => setMonths(e.target.value as 'amanta' | 'purnimanta')} title="Amanta (South/West India) or Purnimanta (North India) month naming">
                <option value="amanta">Amanta</option><option value="purnimanta">Purnimanta</option></select></label>
          </div>
        </nav>
      </header>
      <main>
        {page === 'kundali' && <KundaliPage />}
        {page === 'match' && <MatchPage />}
        {page === 'muhurta' && <MuhurtaPage />}
        {(page === 'day' || page === 'month') && <PanchangPage key={page} mode={page} />}
        {page === 'privacy' && <PrivacyPage />}
      </main>
      <footer className="site-footer">Antariksha is free: no payments, no remedies for sale. · <a href="#privacy">Privacy & your data</a> · Swiss Ephemeris · GeoNames (CC BY 4.0) · For reflection, not certainty.</footer>
    </div>
  );
}

export default function App() { return <SettingsProvider><Shell /></SettingsProvider>; }
