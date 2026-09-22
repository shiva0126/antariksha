import { useEffect, useMemo, useState } from 'react';
import { deleteChatSession, getChart, getReading } from '../../api/client';
import type { ChartResponse, ReadingResponse } from '../../api/types';
import { useT } from '../../i18n';
import { ChatPanel } from '../chat/ChatPanel';
import { chatKey, loadProfiles, newId, saveProfiles, type Profile } from '../common/profiles';
import { AshtakavargaTab } from './AshtakavargaTab';
import { BirthForm, type BirthDetails } from './BirthForm';
import { ChartTab } from './ChartTab';
import { DashaTimeline } from './DashaTimeline';
import { PlanetTable } from './PlanetTable';
import { ProfileBar } from './ProfileBar';
import { ReadingPanel } from './ReadingPanel';
import { ReportView } from './ReportView';
import { StrengthTab } from './StrengthTab';
import { TodayTab } from './TodayTab';

type Tab = 'today' | 'chart' | 'planets' | 'dasha' | 'ashtaka' | 'strength' | 'reading' | 'ask' | 'report';
const tabIds: [Tab, string][] = [['today', 'Today'], ['chart', 'Chart'], ['planets', 'Planets'], ['dasha', 'Dasha'], ['ashtaka', 'Ashtakavarga'], ['strength', 'Strength'], ['reading', 'Reading'], ['ask', 'Ask Antariksha'], ['report', 'Report']];

export function KundaliPage() {
  const t = useT();
  const initial = useMemo(loadProfiles, []);
  const [profiles, setProfiles] = useState<Profile[]>(initial.profiles);
  const [activeId, setActiveId] = useState<string | undefined>(initial.active);
  const [mode, setMode] = useState<'view' | 'edit' | 'new'>(initial.profiles.length ? 'view' : 'new');
  const [chart, setChart] = useState<ChartResponse>();
  const [reading, setReading] = useState<ReadingResponse>();
  const [error, setError] = useState(''), [readingError, setReadingError] = useState('');
  const [busy, setBusy] = useState(false);
  const [tab, setTab] = useState<Tab>(() => (sessionStorage.getItem('antariksha.tab') as Tab) || 'chart');
  const profile = profiles.find(p => p.id === activeId);

  useEffect(() => { try { sessionStorage.setItem('antariksha.tab', tab); } catch { /* ignore */ } }, [tab]);
  useEffect(() => { saveProfiles(profiles, activeId); }, [profiles, activeId]);

  useEffect(() => {
    if (!profile || mode !== 'view') return;
    const ctrl = new AbortController();
    setBusy(true); setError(''); setReadingError(''); setChart(undefined); setReading(undefined);
    getChart(profile, ctrl.signal)
      .then(c => { setChart(c); return getReading(profile, ctrl.signal).then(setReading).catch(e => { if (e.name !== 'AbortError') setReadingError(e.message); }); })
      .catch(e => { if (e.name !== 'AbortError') { setError(e.message); setMode('edit'); } })
      .finally(() => { if (!ctrl.signal.aborted) setBusy(false); });
    return () => ctrl.abort();
    // Recompute only when the chart inputs change, not on every profile edit.
  }, [profile?.id, profile?.date, profile?.time, profile?.lat, profile?.lon, profile?.tz, mode]);

  function submit(b: BirthDetails) {
    if (mode === 'edit' && profile) {
      setProfiles(ps => ps.map(p => (p.id === profile.id ? { ...p, ...b } : p)));
    } else {
      const p: Profile = { ...b, id: newId() };
      setProfiles(ps => [...ps, p]);
      setActiveId(p.id);
    }
    setMode('view'); setTab('chart');
  }

  async function remove(id: string) {
    const p = profiles.find(x => x.id === id);
    if (p) {
      try { const sid = localStorage.getItem(chatKey(p)); if (sid) { await deleteChatSession(sid).catch(() => {}); localStorage.removeItem(chatKey(p)); } } catch { /* ignore */ }
    }
    const rest = profiles.filter(x => x.id !== id);
    setProfiles(rest);
    setActiveId(rest[0]?.id);
    if (!rest.length) setMode('new');
  }

  function print() { setTab('report'); setTimeout(() => window.print(), 600); }

  if (mode !== 'view' || !profile) {
    return (
      <section className="hero">
        <div className="orb" aria-hidden />
        <p className="kicker">A map of the moment you arrived</p>
        <h1>{mode === 'edit' ? 'Edit birth details' : 'The sky remembers.'}</h1>
        <p className="lede">Enter birth details for your Rashi and divisional charts, dashas, Ashtakavarga, today's transits and a grounded reading, then ask questions about your chart. Everything is free.</p>
        <BirthForm key={mode + (profile?.id ?? '')} onSubmit={submit} busy={busy} initial={mode === 'edit' ? profile : undefined} />
        {error && <p role="alert" className="form-error">{error}</p>}
        {profiles.length > 0 && <button className="ghost" onClick={() => setMode('view')}>← Back to {profile?.name || 'my chart'}</button>}
        <p className="hero-foot">Lahiri sidereal · whole-sign houses · Swiss Ephemeris</p>
      </section>
    );
  }

  return (
    <div className="page kundali">
      {!chart ? <div className="card loading-card">{busy ? 'Calculating the chart…' : error}</div> : (
        <>
          <ProfileBar profile={profile} profiles={profiles} chart={chart} reading={reading}
            onEdit={() => setMode('edit')} onSwitch={id => { setActiveId(id); }} onAdd={() => setMode('new')} onDelete={remove} onPrint={print} />
          <nav className="tabs" role="tablist" aria-label="Kundali sections">
            {tabIds.map(([id, label]) => <button key={id} role="tab" aria-selected={tab === id} className={tab === id ? 'active' : ''} onClick={() => setTab(id)}>{t(label)}</button>)}
          </nav>
          {readingError && ['dasha', 'reading', 'ashtaka'].includes(tab) && <p role="alert" className="form-error">Reading unavailable: {readingError}</p>}
          <div role="tabpanel">
            {tab === 'today' && <TodayTab birth={profile} />}
            {tab === 'chart' && <ChartTab chart={chart} birth={profile} reading={reading} onAsk={() => setTab('ask')} />}
            {tab === 'planets' && <PlanetTable chart={chart} facts={reading?.facts} />}
            {tab === 'dasha' && (reading ? <DashaTimeline facts={reading.facts} /> : <div className="card loading-card">Preparing dashas…</div>)}
            {tab === 'ashtaka' && (reading ? <AshtakavargaTab facts={reading.facts} /> : <div className="card loading-card">Preparing Ashtakavarga…</div>)}
            {tab === 'strength' && <StrengthTab birth={profile} />}
            {tab === 'reading' && (reading ? <ReadingPanel data={reading} /> : <div className="card loading-card">Preparing the reading…</div>)}
            {tab === 'ask' && <ChatPanel key={profile.id} birth={profile} name={profile.name} />}
            {tab === 'report' && <ReportView profile={profile} chart={chart} reading={reading} />}
          </div>
        </>
      )}
    </div>
  );
}
