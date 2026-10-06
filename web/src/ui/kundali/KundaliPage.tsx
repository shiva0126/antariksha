import { useEffect, useMemo, useState } from 'react';
import { deleteChatSession, getChart, getReading } from '../../api/client';
import type { ChartResponse, ReadingResponse } from '../../api/types';
import { Button, ButtonLink, Card, Notice, PageHeader, Skeleton, Tabs } from '../../ds';
import { useT } from '../../i18n';
import { go, href, pick } from '../../lib/router';
import { ChatPanel } from '../chat/ChatPanel';
import { chatKey, loadProfiles, newId, saveProfiles, type Profile } from '../common/profiles';
import { MoreReadings } from '../MoreReadings';
import { AshtakavargaTab } from './AshtakavargaTab';
import { BirthForm, type BirthDetails } from './BirthForm';
import { ChartBar } from './ChartBar';
import { ChartTab } from './ChartTab';
import { DashaTimeline } from './DashaTimeline';
import { PlanetTable } from './PlanetTable';
import { ReadingPanel } from './ReadingPanel';
import { ReportView } from './ReportView';
import { StrengthTab } from './StrengthTab';
import { TodayTab } from './TodayTab';
import { TimingTab } from './TimingTab';
import './kundali.css';

const TABS = ['overview', 'chart', 'planets', 'dasha', 'reading', 'ask', 'systems'] as const;
type Tab = typeof TABS[number];

export function KundaliPage({ sub = '' }: { sub?: string }) {
  const t = useT();
  const initial = useMemo(loadProfiles, []);
  const [profiles, setProfiles] = useState<Profile[]>(initial.profiles);
  const [activeId, setActiveId] = useState<string | undefined>(initial.active);
  const [mode, setMode] = useState<'view' | 'edit' | 'new'>(initial.profiles.length ? 'view' : 'new');
  const [chart, setChart] = useState<ChartResponse>();
  const [reading, setReading] = useState<ReadingResponse>();
  const [error, setError] = useState(''), [readingError, setReadingError] = useState('');
  const [busy, setBusy] = useState(false), [report, setReport] = useState(false);
  const tab: Tab = pick(sub, TABS);
  const profile = profiles.find(p => p.id === activeId);

  useEffect(() => { saveProfiles(profiles, activeId); }, [profiles, activeId]);
  useEffect(() => {
    const reload = () => { const s = loadProfiles(); setProfiles(s.profiles); setActiveId(s.active); setMode(s.profiles.length ? 'view' : 'new'); };
    window.addEventListener('astrisk-charts-loaded', reload);
    return () => window.removeEventListener('astrisk-charts-loaded', reload);
  }, []);
  useEffect(() => { if (tab !== 'reading') setReport(false); }, [tab]);

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
    if (mode === 'edit' && profile) setProfiles(ps => ps.map(p => (p.id === profile.id ? { ...p, ...b } : p)));
    else { const p: Profile = { ...b, id: newId() }; setProfiles(ps => [...ps, p]); setActiveId(p.id); }
    setMode('view'); go('kundali', 'chart');
  }

  async function remove(id: string) {
    const p = profiles.find(x => x.id === id);
    if (p) { try { const sid = localStorage.getItem(chatKey(p)); if (sid) { await deleteChatSession(sid).catch(() => {}); localStorage.removeItem(chatKey(p)); } } catch { /* ignore */ } }
    const rest = profiles.filter(x => x.id !== id);
    setProfiles(rest); setActiveId(rest[0]?.id);
    if (!rest.length) setMode('new');
  }

  function print() { go('kundali', 'reading'); setReport(true); setTimeout(() => window.print(), 700); }

  if (tab === 'systems' && (!profile || mode !== 'view')) {
    return <div className="page"><PageHeader kicker="Kundali" title={t('Other systems')} description="Numerology, Western astrology and tarot. These need only your birth date or details you enter here." /><MoreReadings embedded /></div>;
  }
  if (mode !== 'view' || !profile) {
    return (
      <div className="page kundali-start">
        <PageHeader kicker="A map of the moment you arrived" title={mode === 'edit' ? 'Edit birth details' : 'The sky remembers.'}
          description="Enter birth details for your rashi and divisional charts, dashas, Ashtakavarga, today's transits and a grounded reading, then ask questions about your chart. Everything is free." />
        <Card>
          <BirthForm key={mode + (profile?.id ?? '')} onSubmit={submit} busy={busy} initial={mode === 'edit' ? profile : undefined} />
          {error && <Notice tone="danger">{error}</Notice>}
        </Card>
        <div className="ds-row">
          {profiles.length > 0 && <Button variant="ghost" onClick={() => setMode('view')}>← Back to {profile?.name || 'my chart'}</Button>}
          <ButtonLink variant="link" href={href('me', 'charts')}>Load charts saved in your account</ButtonLink>
        </div>
        <p className="ds-muted ds-small">Lahiri sidereal · whole-sign houses · Swiss Ephemeris</p>
      </div>
    );
  }

  const tabs = [{ id: 'overview', label: t('Overview') }, { id: 'chart', label: t('Chart') }, { id: 'planets', label: t('Planets') }, { id: 'dasha', label: t('Timing') }, { id: 'reading', label: t('Reading') }, { id: 'ask', label: t('Ask Astrisk') }, { id: 'systems', label: t('Other systems') }] as { id: Tab; label: string }[];
  return (
    <div className="page kundali">
      {!chart ? <Card>{busy ? <Skeleton lines={4} /> : <Notice tone="danger">{error}</Notice>}</Card> : (
        <>
          <ChartBar profile={profile} profiles={profiles} chart={chart} reading={reading}
            onEdit={() => setMode('edit')} onSwitch={setActiveId} onAdd={() => setMode('new')} onDelete={id => void remove(id)} onPrint={print} />
          <Tabs label="Kundali sections" base="kundali" active={tab} items={tabs} />
          {readingError && ['dasha', 'reading', 'planets'].includes(tab) && <Notice tone="danger">Reading unavailable: {readingError}</Notice>}
          <div role="tabpanel" className="ds-stack">
            {tab === 'overview' && <TodayTab birth={profile} />}
            {tab === 'chart' && <ChartTab chart={chart} birth={profile} reading={reading} onAsk={() => go('kundali', 'ask')} />}
            {tab === 'planets' && <>
              <PlanetTable chart={chart} facts={reading?.facts} />
              <StrengthTab birth={profile} />
              {reading ? <AshtakavargaTab facts={reading.facts} /> : <Card><Skeleton /></Card>}
            </>}
            {tab === 'dasha' && <><TimingTab birth={profile} />{reading ? <DashaTimeline facts={reading.facts} /> : <Card><Skeleton /></Card>}</>}
            {tab === 'reading' && (report ? <><div className="ds-row no-print"><Button variant="ghost" onClick={() => setReport(false)}>← Back to the reading</Button><Button variant="primary" onClick={() => window.print()}>Print / save as PDF</Button></div><ReportView profile={profile} chart={chart} reading={reading} /></>
              : reading ? <><div className="ds-row"><Button onClick={() => setReport(true)}>{t('Report')}: printable kundali</Button></div><ReadingPanel data={reading} /></> : <Card><Skeleton lines={5} /></Card>)}
            {tab === 'ask' && <ChatPanel key={profile.id} birth={profile} name={profile.name} />}
            {tab === 'systems' && <MoreReadings embedded />}
          </div>
        </>
      )}
    </div>
  );
}
