import { useEffect, useState } from 'react';
import { useT } from '../../i18n';
import { FamilyBiodata } from '../community/Biodata';
import { FamilyAssistance } from '../community/FamilyAssistance';
import { memberAPI } from '../community/shared';
import { Discover } from './Discover';
import { Interests } from './Interests';
import { ProfileEditor } from './ProfileEditor';
import { Safety } from './Safety';
import type { MyProfile } from './types';
import './matrimony.css';

type Tab = 'discover' | 'interests' | 'profile' | 'family' | 'safety';
const TABS: [Tab, string][] = [['discover', 'Discover'], ['interests', 'Interests'], ['profile', 'My matrimony profile'], ['family', 'Family'], ['safety', 'Safety']];

/** Matrimony: its own section, reached from the main navigation. Sub-pages
 *  are addressed as #matrimony/<tab>, so alerts can deep-link. */
export function MatrimonyPage() {
  const t = useT();
  const fromHash = (): Tab => { const s = location.hash.split('/')[1] as Tab; return TABS.some(([k]) => k === s) ? s : 'discover'; };
  const [tab, setTab] = useState<Tab>(fromHash);
  const [profile, setProfile] = useState<MyProfile | null>(), [community, setCommunity] = useState<boolean>(), [pending, setPending] = useState(0);
  async function load() {
    const [rows, settings, interests] = await Promise.all([memberAPI<MyProfile[]>('/api/matrimony/me'), memberAPI<{ community: boolean }[]>('/api/community/settings').then(r => r[0] ?? { community: false }).catch(() => ({ community: false })), memberAPI<{ outgoing: boolean; status: string }[]>('/api/matrimony/interests').catch(() => [])]);
    setProfile(rows[0] ?? null); setCommunity(settings.community);
    setPending(interests.filter(i => !i.outgoing && i.status === 'pending').length);
  }
  // Reload on every tab change so counts (e.g. new interests) stay current.
  useEffect(() => { void load(); }, [tab]);
  useEffect(() => { const f = () => setTab(fromHash()); window.addEventListener('hashchange', f); return () => window.removeEventListener('hashchange', f); }, []);
  const go = (k: Tab) => { history.replaceState(null, '', '#matrimony/' + k); setTab(k); };
  const ready = profile !== undefined;
  return (
    <div className="page matrimony">
      <header className="page-head">
        <div><p className="kicker">Matrimony · horoscope-aware</p><h1>{t('Matrimony')}</h1>
          <p className="muted">Profiles you can filter, with each person's guna score and the reasons they were suggested. Free, private by default, and you choose what to share.</p></div>
      </header>
      {ready && community === false && <div className="card notice"><p>To use matrimony, first turn on <b>Community</b> with your adult date of birth in <a href="#community">Community → Profile &amp; interests</a>.</p></div>}
      {ready && !profile && community !== false && tab !== 'profile' && <div className="card notice"><p>Start by creating your matrimony profile. It stays private until you make it discoverable. <button className="primary" onClick={() => go('profile')}>Create my profile</button></p></div>}
      <nav className="tabs mat-tabs" role="tablist" aria-label="Matrimony sections">
        {TABS.map(([k, l]) => <button key={k} role="tab" aria-selected={tab === k} className={tab === k ? 'active' : ''} onClick={() => go(k)}>{t(l)}{k === 'interests' && pending > 0 && <span className="badge-count" aria-label={`${pending} new`}>{pending}</span>}</button>)}
      </nav>
      {tab === 'discover' && ready && <Discover key={JSON.stringify(profile?.saved_search ?? {})} saved={profile?.saved_search ?? {}} horoscopeOn={!!profile?.horoscope_visible} />}
      {tab === 'interests' && <Interests />}
      {tab === 'profile' && <ProfileEditor onSaved={() => void load()} />}
      {tab === 'family' && <><FamilyBiodata onImported={load} /><FamilyAssistance /></>}
      {tab === 'safety' && <Safety />}
      <p className="disclaimer">Suggestions reflect stated preferences and, if both members opt in, traditional horoscope tables. They are not a prediction of relationship success.</p>
    </div>
  );
}
