import type { ChartFacts, ChartResponse, Graha } from '../../api/types';
import { degreeLabel, houseNumber } from '../../astro/houses';
import { grahaColor, grahaGlyph } from '../../astro/rashi';
import { grahaEnglish, titleCase } from '../../astro/format';

export function GrahaCard({ graha, chart, facts, meaning, onClose }: { graha: Graha; chart: ChartResponse; facts?: ChartFacts; meaning?: string; onClose: () => void }) {
  const d = facts?.dignities[graha.id];
  const rows: [string, string][] = [
    ['Rashi', `${graha.rashi} · ${degreeLabel(graha.rashi_degree)}`],
    ['House', String(houseNumber(graha.longitude, chart.ascendant.longitude))],
    ['Nakshatra', `${graha.nakshatra} · pada ${graha.nakshatra_pada}`],
    ['Longitude', `${graha.longitude.toFixed(3)}°`],
    ['Motion', graha.id === 'rahu' || graha.id === 'ketu' ? 'Lunar node' : `${graha.retrograde ? 'Retrograde ℞' : 'Direct'} · ${graha.speed.toFixed(3)}°/day`],
  ];
  if (d) rows.push(['Dignity', titleCase(d.state) + (d.neecha_bhanga ? ' (neecha bhanga)' : '') + (facts?.combustion[graha.id] ? ' · combust' : '')]);
  return (
    <aside className="card graha-card" aria-label={`${graha.name} details`}>
      <button className="icon-button" onClick={onClose} aria-label="Close planet details">×</button>
      <div className="graha-head">
        <span className="graha-glyph" style={{ color: grahaColor[graha.id] }} aria-hidden>{grahaGlyph[graha.id]}</span>
        <div><p className="kicker">Graha</p><h3>{graha.name} <small>{grahaEnglish(graha.id)}</small></h3></div>
      </div>
      <dl className="kv">{rows.map(([k, v]) => <div key={k}><dt>{k}</dt><dd>{v}</dd></div>)}</dl>
      {meaning && <p className="graha-meaning">{meaning}</p>}
    </aside>
  );
}
