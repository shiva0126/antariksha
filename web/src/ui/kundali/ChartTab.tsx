import { lazy, Suspense, useEffect, useState } from 'react';
import { getVarga } from '../../api/client';
import type { ChartInput, ChartResponse, Graha, ReadingResponse } from '../../api/types';
import { CircularWheel } from '../../chart2d/CircularWheel';
import { NorthIndian } from '../../chart2d/NorthIndian';
import { SouthIndian } from '../../chart2d/SouthIndian';
import { GrahaCard } from './GrahaCard';

const Dome = lazy(() => import('../../scene/CelestialDome'));
type Layout = 'south' | 'north' | 'circular' | 'dome';
const labels: Record<Layout, string> = { south: 'South Indian', north: 'North Indian', circular: 'Circular', dome: 'Celestial dome' };
const vargas: [number, string, string][] = [[1, 'D1 Rashi', 'body and overall life'], [9, 'D9 Navamsha', 'marriage and dharma'], [10, 'D10 Dashamsha', 'career'], [2, 'D2 Hora', 'wealth'], [3, 'D3 Drekkana', 'siblings'], [7, 'D7 Saptamsha', 'children'], [12, 'D12 Dwadashamsha', 'parents']];

function webgl() { try { const c = document.createElement('canvas'); return !!(c.getContext('webgl2') || c.getContext('webgl')); } catch { return false; } }

export function ChartTab({ chart: natal, birth, reading, onAsk }: { chart: ChartResponse; birth: ChartInput; reading?: ReadingResponse; onAsk: () => void }) {
  const [varga, setVarga] = useState(1);
  const [vchart, setVchart] = useState<ChartResponse>();
  const [vError, setVError] = useState('');
  useEffect(() => {
    if (varga === 1) { setVchart(undefined); return; }
    const ctrl = new AbortController();
    setVchart(undefined); setVError('');
    getVarga(birth, varga, ctrl.signal).then(r => setVchart(r.chart)).catch(e => { if (e.name !== 'AbortError') setVError(e.message); });
    return () => ctrl.abort();
  }, [varga, birth]);
  const chart = varga === 1 ? natal : vchart ?? natal;
  const [layout, setLayout] = useState<Layout>(() => (localStorage.getItem('antariksha.layout') as Layout) || 'south');
  const [selected, setSelected] = useState<Graha>();
  const [can3d] = useState(webgl);
  const choose = (l: Layout) => { setLayout(l); try { localStorage.setItem('antariksha.layout', l); } catch { /* ignore */ } };
  const facts = reading?.facts;
  const meaning = selected ? reading?.reading.grahas.find(g => g.graha.startsWith(selected.name))?.meaning : undefined;
  const yogas = facts?.yogas ?? [];
  return (
    <div className="chart-tab">
      <div className="card chart-card">
        <div className="segmented" role="tablist" aria-label="Chart style">
          {(Object.keys(labels) as Layout[]).map(l => (
            <button key={l} role="tab" aria-selected={layout === l} className={layout === l ? 'active' : ''} disabled={l === 'dome' && !can3d} onClick={() => choose(l)}>{labels[l]}</button>
          ))}
        </div>
        <label className="field compact varga-select"><span>Divisional chart</span>
          <select value={varga} onChange={e => setVarga(Number(e.target.value))} aria-label="Divisional chart">
            {vargas.map(([n, name, theme]) => <option key={n} value={n}>{name} · {theme}</option>)}
          </select>
        </label>
        {vError && <p className="form-error">{vError}</p>}
        <div className="chart-stage">
          {layout === 'dome' && can3d && varga === 1 ? (
            <Suspense fallback={<div className="loading">Assembling the heavens…</div>}><Dome chart={chart} onSelect={setSelected} /></Suspense>
          ) : layout === 'north' ? <NorthIndian chart={chart} onSelect={setSelected} selected={selected?.id} />
            : layout === 'circular' ? <CircularWheel chart={chart} onSelect={setSelected} selected={selected?.id} />
              : <SouthIndian chart={chart} onSelect={setSelected} selected={selected?.id} title={varga === 1 ? 'Rashi · D1' : vargas.find(v => v[0] === varga)?.[1].replace(' ', ' · ')} />}
        </div>
        <p className="muted small chart-note">{vargas.find(v => v[0] === varga)?.[1]} · Lahiri sidereal · whole-sign houses. {varga === 1 ? 'Tap a planet for details.' : 'Divisional positions are shown within their division; tap a planet for its natal details.'}</p>
      </div>
      <div className="chart-side">
        {selected ? <GrahaCard graha={natal.grahas.find(g => g.id === selected.id) ?? selected} chart={natal} facts={facts} meaning={meaning} onClose={() => setSelected(undefined)} /> : (
          <div className="card">
            <h3>At a glance</h3>
            {!reading ? <p className="muted">Preparing chart facts…</p> : (
              <>
                <p>{reading.reading.summary}</p>
                <h4>Yogas</h4>
                {yogas.length === 0 ? <p className="muted small">None from the engine's catalogue.</p> : (
                  <ul className="tag-list">{yogas.map(y => <li key={y.name} className={'tag tag-' + y.type}>{y.name}</li>)}</ul>
                )}
              </>
            )}
            <button className="primary block" onClick={onAsk}>Ask a question about this chart</button>
          </div>
        )}
      </div>
    </div>
  );
}
