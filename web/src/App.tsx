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
import { AdminPage } from './ui/AdminPage';
import { LoginPage } from './ui/LoginPage';
import { EmailLinkGate } from './ui/EmailRecovery';
import { initializeAccountCharts, setProfileOwner } from './ui/common/profiles';

type Page = 'kundali' | 'match' | 'muhurta' | 'day' | 'month' | 'privacy' | 'account' | 'community' | 'readings' | 'admin';
const pages: [Page, string][] = [['kundali', 'Kundali'], ['match', 'Matching'], ['muhurta', 'Muhurta'], ['day', 'Daily Panchang'], ['month', 'Hindu Calendar'], ['readings', 'More readings']];
const valid = new Set<string>(['kundali', 'match', 'muhurta', 'day', 'month', 'privacy', 'account', 'community', 'readings','admin']);
const fromHash = (): Page => { const p = location.hash.slice(1); return (valid.has(p) ? p : 'kundali') as Page; };

function Shell({role}:{role:string}) {
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
        <button className="brand" onClick={() => go('kundali')} aria-label="Astrisk home"><span className="brand-mark" aria-hidden>✦</span><b>Astrisk</b></button>
        <button className="menu-toggle" aria-expanded={menu} aria-controls="main-nav" onClick={() => setMenu(!menu)}>Menu</button>
        <nav id="main-nav" className={'main-nav' + (menu ? ' open' : '')} aria-label="Main navigation">
          {pages.map(([id, label]) => <button key={id} className={page === id ? 'active' : ''} aria-current={page === id ? 'page' : undefined} onClick={() => go(id)}>{t(label)}</button>)}
          <button onClick={() => go('account')} className={page === 'account' ? 'active' : ''}>My profile</button>
          <button onClick={() => go('community')} className={page === 'community' ? 'active' : ''}>Community</button>
          {role==='superadmin'&&<button onClick={()=>go('admin')} className={page==='admin'?'active':''}>Superadmin</button>}
          <div className="settings">
            <label><span className="sr-only">{t('Language')}</span>
              <select aria-label="Language" value={lang} onChange={e => setLang(e.target.value as Lang)}>{LANGUAGES.map(([k, label]) => <option key={k} value={k}>{label}</option>)}</select></label>
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
        {page === 'account' && <AccountPage />}
        {page === 'community' && <CommunityPage />}
        {page === 'readings' && <MoreReadings />}
        {page === 'admin' && (role==='superadmin'?<AdminPage/>:<div className="page"><h1>Admin access required</h1></div>)}
      </main>
      <footer className="site-footer">Astrisk is free: no payments, no remedies for sale. · <a href="#privacy">Privacy & your data</a> · Swiss Ephemeris · GeoNames (CC BY 4.0) · For reflection, not certainty.</footer>
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
  const check=async()=>{const run=++sequence.current;const response=await fetch('/api/me',{credentials:'same-origin'});if(run!==sequence.current)return;if(response.ok){const me=await response.json();if(run!==sequence.current)return;if(identity.current!==me.id){setLoading(true);setSignedIn(false);}identity.current=me.id;setProfileOwner(me.id,{date:me.birth_date,time:me.birth_time?.slice(0,5)});await initializeAccountCharts().catch(()=>{});if(run!==sequence.current)return;setAccountID(me.id);setRole(me.role||'member');}else{identity.current='';setRole('member');setProfileOwner('signed-out');}setSignedIn(response.ok);setLoading(false);};
  useEffect(()=>{void check().catch(()=>setLoading(false));const expired=()=>{sequence.current++;identity.current='';setProfileOwner('signed-out');setSignedIn(false);setLoading(false);};const refresh=()=>{void check().catch(()=>{setSignedIn(false);setLoading(false);});};window.addEventListener('antariksha-signed-out',expired);window.addEventListener('focus',refresh);return()=>{sequence.current++;window.removeEventListener('antariksha-signed-out',expired);window.removeEventListener('focus',refresh);};},[]);
  if(loading)return <main className="login-screen"><p>Opening Astrisk…</p></main>;
  if(!signedIn)return <LoginPage onSignedIn={check}/>;
  return <SettingsProvider key={accountID}><Shell role={role}/></SettingsProvider>;
}
