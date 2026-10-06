import { useEffect, useState } from 'react';
import { useT } from '../../i18n';
import { Notice, PageHeader, Tabs } from '../../ds';
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
      <PageHeader kicker="Matrimony · horoscope-aware" title={t('Matrimony')} compactOnMobile description="Profiles you can filter, with each person's guna score and why they were suggested. Free, private by default, and you choose what to share." />
      {ready && community === false && <Notice tone="warning">To use matrimony, first turn on your community profile with your adult date of birth in <a href="#me/profile">Me → Profile</a>.</Notice>}
      {ready && !profile && community !== false && tab !== 'profile' && <Notice>Start with your biodata. It stays private until you make it discoverable. <a href="#matrimony/profile">Create my biodata →</a></Notice>}
      <Tabs label="Matrimony sections" base="matrimony" active={tab} items={TABS.map(([k, l]) => ({ id: k, label: t(l), count: k === 'interests' ? pending : 0 }))} />
      {tab === 'discover' && ready && <Discover key={JSON.stringify(profile?.saved_search ?? {})} saved={profile?.saved_search ?? {}} horoscopeOn={!!profile?.horoscope_visible} />}
      {tab === 'interests' && <Interests />}
      {tab === 'profile' && <ProfileEditor onSaved={() => void load()} />}
      {tab === 'family' && <><FamilyBiodata onImported={load} /><FamilyAssistance /></>}
      {tab === 'safety' && <Safety />}
      <p className="disclaimer">Suggestions reflect stated preferences and, if both members opt in, traditional horoscope tables. They are not a prediction of relationship success.</p>
    </div>
  );
}
