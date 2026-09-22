import { useEffect, useState } from 'react';
import { getChart, getReading } from '../../api/client';
import type { ChartResponse, ReadingResponse } from '../../api/types';
import { ChatPanel } from '../chat/ChatPanel';
import { BirthForm, type BirthDetails } from './BirthForm';
import { ChartTab } from './ChartTab';
import { DashaTimeline } from './DashaTimeline';
import { PlanetTable } from './PlanetTable';
import { ProfileBar } from './ProfileBar';
import { ReadingPanel } from './ReadingPanel';

type Tab = 'chart' | 'planets' | 'dasha' | 'reading' | 'ask';
const tabs: [Tab, string][] = [['chart', 'Chart'], ['planets', 'Planets'], ['dasha', 'Dasha'], ['reading', 'Reading'], ['ask', 'Ask Antariksha']];
const STORE = 'antariksha.birth';

function stored(): BirthDetails | undefined {
  try { const v = localStorage.getItem(STORE); return v ? JSON.parse(v) : undefined; } catch { return undefined; }
}

export function KundaliPage() {
  const [birth, setBirth] = useState<BirthDetails | undefined>(stored);
  const [editing, setEditing] = useState(!birth);
  const [chart, setChart] = useState<ChartResponse>();
  const [reading, setReading] = useState<ReadingResponse>();
  const [error, setError] = useState(''), [readingError, setReadingError] = useState('');
  const [busy, setBusy] = useState(false);
  const [tab, setTab] = useState<Tab>(() => (sessionStorage.getItem('antariksha.tab') as Tab) || 'chart');

  useEffect(() => { try { sessionStorage.setItem('antariksha.tab', tab); } catch { /* ignore */ } }, [tab]);

  useEffect(() => {
    if (!birth || editing) return;
    const ctrl = new AbortController();
    setBusy(true); setError(''); setReadingError(''); setChart(undefined); setReading(undefined);
    getChart(birth, ctrl.signal)
      .then(c => { setChart(c); return getReading(birth, ctrl.signal).then(setReading).catch(e => { if (e.name !== 'AbortError') setReadingError(e.message); }); })
      .catch(e => { if (e.name !== 'AbortError') { setError(e.message); setEditing(true); } })
      .finally(() => { if (!ctrl.signal.aborted) setBusy(false); });
    return () => ctrl.abort();
  }, [birth, editing]);

  function submit(b: BirthDetails) {
    try { localStorage.setItem(STORE, JSON.stringify(b)); } catch { /* ignore */ }
    setBirth(b); setEditing(false); setTab('chart');
  }

  if (editing || !birth) {
    return (
      <section className="hero">
        <div className="orb" aria-hidden />
        <p className="kicker">A map of the moment you arrived</p>
        <h1>The sky remembers.</h1>
        <p className="lede">Enter birth details for your Rashi kundali, planetary degrees, nakshatras, Vimshottari dashas and a grounded reading — then ask questions about your chart.</p>
        <BirthForm onSubmit={submit} busy={busy} initial={birth} />
        {error && <p role="alert" className="form-error">{error}</p>}
        {birth && <button className="ghost" onClick={() => setEditing(false)}>Back to {birth.name || 'my chart'}</button>}
        <p className="hero-foot">Lahiri sidereal · whole-sign houses · Swiss Ephemeris</p>
      </section>
    );
  }

  return (
    <div className="page kundali">
      {!chart ? <div className="card loading-card">{busy ? 'Calculating the chart…' : error}</div> : (
        <>
          <ProfileBar birth={birth} chart={chart} reading={reading} onEdit={() => setEditing(true)} />
          <nav className="tabs" role="tablist" aria-label="Kundali sections">
            {tabs.map(([id, label]) => (
              <button key={id} role="tab" aria-selected={tab === id} className={tab === id ? 'active' : ''} onClick={() => setTab(id)}>{label}</button>
            ))}
          </nav>
          {readingError && tab !== 'chart' && tab !== 'planets' && tab !== 'ask' && <p role="alert" className="form-error">Reading unavailable: {readingError}</p>}
          <div role="tabpanel">
            {tab === 'chart' && <ChartTab chart={chart} reading={reading} onAsk={() => setTab('ask')} />}
            {tab === 'planets' && <PlanetTable chart={chart} facts={reading?.facts} />}
            {tab === 'dasha' && (reading ? <DashaTimeline facts={reading.facts} /> : <div className="card loading-card">Preparing dashas…</div>)}
            {tab === 'reading' && (reading ? <ReadingPanel data={reading} /> : <div className="card loading-card">Preparing the reading…</div>)}
            {tab === 'ask' && <ChatPanel birth={birth} name={birth.name} />}
          </div>
        </>
      )}
    </div>
  );
}
