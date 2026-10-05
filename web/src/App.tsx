import { useEffect, useRef, useState } from 'react';
import { LANGUAGES, SettingsProvider, useSettings, useT, type Lang } from './i18n';
import { KundaliPage } from './ui/kundali/KundaliPage';
import { MatchPage } from './ui/match/MatchPage';
import { MuhurtaPage } from './ui/muhurta/MuhurtaPage';
import { PanchangPage } from './ui/panchang/PanchangPage';
import { PrivacyPage } from './ui/privacy/PrivacyPage';
import { AccountPage } from './ui/AccountPage';
import { CommunityPage } from './ui/community/CommunityPage';
import { MoreReadings } from './ui/MoreReadings';
import { MatrimonyPage } from './ui/matrimony/MatrimonyPage';
import { BiodataPrint } from './ui/matrimony/BiodataPrint';
import { AdminPage } from './ui/AdminPage';
import { LoginPage } from './ui/LoginPage';
import { EmailLinkGate } from './ui/EmailRecovery';
import { ErrorBoundary } from './ui/common/ErrorBoundary';
import { initializeAccountCharts, setProfileOwner } from './ui/common/profiles';

type Page = 'kundali' | 'match' | 'muhurta' | 'day' | 'month' | 'privacy' | 'account' | 'community' | 'readings' | 'admin' | 'matrimony' | 'biodata';
// Five main sections; everything else lives under "More".
const primary: [Page, string, string][] = [['kundali', 'Kundali', '✦'], ['day', 'Panchang', '☾'], ['match', 'Matching', '⚭'], ['matrimony', 'Matrimony', '♡'], ['community', 'Community', '❖']];
const secondary: [Page, string][] = [['muhurta', 'Muhurta'], ['month', 'Hindu Calendar'], ['readings', 'More readings'], ['account', 'My profile'], ['privacy', 'Privacy']];
const valid = new Set<string>(['kundali', 'match', 'muhurta', 'day', 'month', 'privacy', 'account', 'community', 'readings', 'admin', 'matrimony', 'biodata']);
const fromHash = (): Page => { const p = location.hash.slice(1).split('/')[0]; return (valid.has(p) ? p : 'kundali') as Page; };
const section = (p: Page): Page => p === 'month' ? 'day' : p === 'biodata' ? 'matrimony' : p;

function Shell({role}:{role:string}) {
  const t = useT();
  const { lang, setLang, months, setMonths } = useSettings();
  const [page, setPage] = useState<Page>(fromHash);
  const [more, setMore] = useState(false);
  const moreRef = useRef<HTMLDivElement>(null);
  useEffect(() => {
    const changed = () => { setPage(fromHash()); setMore(false); window.scrollTo({ top: 0 }); };
    window.addEventListener('hashchange', changed);
    const close = (e: MouseEvent) => { if (!moreRef.current?.contains(e.target as Node)) setMore(false); };
    document.addEventListener('mousedown', close);
    return () => { window.removeEventListener('hashchange', changed); document.removeEventListener('mousedown', close); };
  }, []);
  function go(p: Page) { location.hash = p; setPage(p); setMore(false); window.scrollTo({ top: 0 }); }
  const current = section(page);
  const extra = role === 'superadmin' ? [...secondary, ['admin', 'Superadmin'] as [Page, string]] : secondary;
  return (
    <div className="app">
      <header className="site-header">
        <button className="brand" onClick={() => go('kundali')} aria-label="Astrisk home"><span className="brand-mark" aria-hidden>✦</span><b>Astrisk</b></button>
        <nav id="main-nav" className="main-nav" aria-label="Main navigation">
          {primary.map(([id, label]) => <button key={id} className={current === id ? 'active' : ''} aria-current={current === id ? 'page' : undefined} onClick={() => go(id)}>{t(label)}</button>)}
        </nav>
        <div className="header-tools">
          <div className="more-menu" ref={moreRef}>
            <button className={'more-toggle' + (extra.some(([id]) => id === page) ? ' active' : '')} aria-expanded={more} aria-haspopup="menu" onClick={() => setMore(!more)}>{t('More')} ▾</button>
            {more && <div className="more-panel" role="menu">
              {extra.map(([id, label]) => <button key={id} role="menuitem" className={page === id ? 'active' : ''} onClick={() => go(id)}>{t(label)}</button>)}
              <div className="settings">
                <label><span>{t('Language')}</span>
                  <select aria-label="Language" value={lang} onChange={e => setLang(e.target.value as Lang)}>{LANGUAGES.map(([k, label]) => <option key={k} value={k}>{label}</option>)}</select></label>
                <label><span>{t('Months')}</span>
                  <select aria-label="Month system" value={months} onChange={e => setMonths(e.target.value as 'amanta' | 'purnimanta')} title="Amanta (South/West India) or Purnimanta (North India) month naming">
                    <option value="amanta">Amanta</option><option value="purnimanta">Purnimanta</option></select></label>
              </div>
            </div>}
          </div>
        </div>
      </header>
      {(page === 'day' || page === 'month') && (
        <div className="page-switch" role="tablist" aria-label="Panchang view">
          <button role="tab" aria-selected={page === 'day'} className={page === 'day' ? 'active' : ''} onClick={() => go('day')}>{t('Daily Panchang')}</button>
          <button role="tab" aria-selected={page === 'month'} className={page === 'month' ? 'active' : ''} onClick={() => go('month')}>{t('Hindu Calendar')}</button>
        </div>
      )}
      <main><ErrorBoundary key={page}>
        {page === 'kundali' && <KundaliPage />}
        {page === 'match' && <MatchPage />}
        {page === 'muhurta' && <MuhurtaPage />}
        {(page === 'day' || page === 'month') && <PanchangPage key={page} mode={page} />}
        {page === 'privacy' && <PrivacyPage />}
        {page === 'account' && <AccountPage />}
        {page === 'community' && <CommunityPage />}
        {page === 'matrimony' && <MatrimonyPage />}
        {page === 'biodata' && <BiodataPrint />}
        {page === 'readings' && <MoreReadings />}
        {page === 'admin' && (role==='superadmin'?<AdminPage/>:<div className="page"><h1>Admin access required</h1></div>)}
      </ErrorBoundary></main>
      <footer className="site-footer">Astrisk is free: no payments, no remedies for sale. · <a href="#privacy">Privacy & your data</a> · Swiss Ephemeris · GeoNames (CC BY 4.0) · For reflection, not certainty.</footer>
      <nav className="tab-bar" aria-label="Sections">
        {primary.map(([id, label, icon]) => <button key={id} className={current === id ? 'active' : ''} aria-current={current === id ? 'page' : undefined} onClick={() => go(id)}><span aria-hidden>{icon}</span>{t(label)}</button>)}
      </nav>
    </div>
  );
}

export default function App() {
 return <EmailLinkGate><AuthenticatedApp/></EmailLinkGate>;
}
function AuthenticatedApp() {
  const [signedIn,setSignedIn]=useState(false),[loading,setLoading]=useState(true);
  const [accountID,setAccountID]=useState('');
  const [role,setRole]=useState('member');
  const sequence=useRef(0),identity=useRef('');
  const check=async()=>{const run=++sequence.current;const response=await fetch('/api/me',{credentials:'same-origin'});if(run!==sequence.current)return;if(response.ok){const me=await response.json();if(run!==sequence.current)return;if(identity.current!==me.id){setLoading(true);setSignedIn(false);}identity.current=me.id;setProfileOwner(me.id,{date:me.birth_date,time:me.birth_time?.slice(0,5),place:me.birth_place??undefined});await initializeAccountCharts().catch(()=>{});if(run!==sequence.current)return;setAccountID(me.id);setRole(me.role||'member');}else{identity.current='';setRole('member');setProfileOwner('signed-out');}setSignedIn(response.ok);setLoading(false);};
  useEffect(()=>{void check().catch(()=>setLoading(false));const expired=()=>{sequence.current++;identity.current='';setProfileOwner('signed-out');setSignedIn(false);setLoading(false);};const refresh=()=>{void check().catch(()=>{setSignedIn(false);setLoading(false);});};window.addEventListener('antariksha-signed-out',expired);window.addEventListener('focus',refresh);return()=>{sequence.current++;window.removeEventListener('antariksha-signed-out',expired);window.removeEventListener('focus',refresh);};},[]);
  if(loading)return <main className="login-screen"><p>Opening Astrisk…</p></main>;
  if(!signedIn)return <LoginPage onSignedIn={check}/>;
  return <SettingsProvider key={accountID}><Shell role={role}/></SettingsProvider>;
}
