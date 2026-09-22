import { lazy, Suspense, useState } from 'react';
import type { ChartResponse, Graha, ReadingResponse } from '../../api/types';
import { CircularWheel } from '../../chart2d/CircularWheel';
import { NorthIndian } from '../../chart2d/NorthIndian';
import { SouthIndian } from '../../chart2d/SouthIndian';
import { GrahaCard } from './GrahaCard';

const Dome = lazy(() => import('../../scene/CelestialDome'));
type Layout = 'south' | 'north' | 'circular' | 'dome';
const labels: Record<Layout, string> = { south: 'South Indian', north: 'North Indian', circular: 'Circular', dome: 'Celestial dome' };

function webgl() { try { const c = document.createElement('canvas'); return !!(c.getContext('webgl2') || c.getContext('webgl')); } catch { return false; } }

export function ChartTab({ chart, reading, onAsk }: { chart: ChartResponse; reading?: ReadingResponse; onAsk: () => void }) {
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
        <div className="chart-stage">
          {layout === 'dome' && can3d ? (
            <Suspense fallback={<div className="loading">Assembling the heavens…</div>}><Dome chart={chart} onSelect={setSelected} /></Suspense>
          ) : layout === 'north' ? <NorthIndian chart={chart} onSelect={setSelected} selected={selected?.id} />
            : layout === 'circular' ? <CircularWheel chart={chart} onSelect={setSelected} selected={selected?.id} />
              : <SouthIndian chart={chart} onSelect={setSelected} selected={selected?.id} />}
        </div>
        <p className="muted small chart-note">Rashi (D1) · Lahiri sidereal · whole-sign houses. Tap a planet for details.</p>
      </div>
      <div className="chart-side">
        {selected ? <GrahaCard graha={selected} chart={chart} facts={facts} meaning={meaning} onClose={() => setSelected(undefined)} /> : (
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
