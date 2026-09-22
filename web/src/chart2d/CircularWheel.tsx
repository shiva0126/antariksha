import { grahaColor, grahaGlyph, rashis } from '../astro/rashi';
import { wheelPoint } from '../astro/mapping';
import { houseNumber } from '../astro/houses';
import { type ChartProps, keyAct } from './common';

export function CircularWheel({ chart, onSelect, selected }: ChartProps) {
  const c = 310, sorted = [...chart.grahas].sort((a, b) => a.longitude - b.longitude);
  const lanes = new Map<string, number>();
  for (const g of sorted) {
    let lane = 0;
    const near = (p: typeof g) => Math.min(Math.abs(g.longitude - p.longitude), 360 - Math.abs(g.longitude - p.longitude)) < 12;
    while (sorted.some(p => lanes.get(p.id) === lane && near(p))) lane++;
    lanes.set(g.id, lane);
  }
  const lagna = wheelPoint(chart.ascendant.longitude, 292, c, c);
  return (
    <svg className="chart-svg" viewBox="0 0 620 620" role="img" aria-label="Kundali wheel: exact sidereal longitudes">
      <circle cx={c} cy={c} r="300" className="c-frame" />
      <circle cx={c} cy={c} r="250" className="c-lines" fill="none" />
      {rashis.map((r, i) => {
        const a = wheelPoint(i * 30, 300, c, c), b = wheelPoint(i * 30, 76, c, c), p = wheelPoint(i * 30 + 15, 275, c, c), h = wheelPoint(i * 30 + 15, 92, c, c);
        return (
          <g key={r[0]}>
            <line x1={a.x} y1={a.y} x2={b.x} y2={b.y} className="c-lines" />
            <text x={p.x} y={p.y - 3} textAnchor="middle" className="c-glyph">{r[2]}</text>
            <text x={p.x} y={p.y + 13} textAnchor="middle" className="c-sub">{r[0]}</text>
            <text x={h.x} y={h.y + 4} textAnchor="middle" className="c-house">{houseNumber(i * 30, chart.ascendant.longitude)}</text>
          </g>
        );
      })}
      {Array.from({ length: 108 }, (_, i) => {
        const a = wheelPoint((i * 360) / 108, 250, c, c), b = wheelPoint((i * 360) / 108, i % 4 ? 246 : 240, c, c);
        return <line key={i} x1={a.x} y1={a.y} x2={b.x} y2={b.y} className="c-tick" />;
      })}
      <line x1={c} y1={c} x2={lagna.x} y2={lagna.y} className="c-lagna-mark" strokeDasharray="4 5" />
      {sorted.map(g => {
        const lane = lanes.get(g.id) ?? 0, p = wheelPoint(g.longitude, 222 - lane * 30, c, c), a = wheelPoint(g.longitude, 249, c, c);
        return (
          <g key={g.id} className={'planet' + (selected === g.id ? ' is-selected' : '')} data-longitude={g.longitude} role="button" tabIndex={0}
            aria-label={`${g.name} ${g.rashi} ${g.rashi_degree.toFixed(2)} degrees`} onClick={() => onSelect?.(g)} onKeyDown={keyAct(() => onSelect?.(g))}>
            <title>{`${g.name} · ${g.rashi} ${g.rashi_degree.toFixed(2)}° · House ${houseNumber(g.longitude, chart.ascendant.longitude)}`}</title>
            <line x1={a.x} y1={a.y} x2={p.x} y2={p.y} stroke={grahaColor[g.id]} opacity=".5" />
            <circle cx={p.x} cy={p.y} r="14" className="c-planet-disc" stroke={grahaColor[g.id]} />
            <text x={p.x} y={p.y + 1} fill={grahaColor[g.id]} textAnchor="middle" dominantBaseline="middle" fontSize="18">{grahaGlyph[g.id]}</text>
            <text x={p.x} y={p.y + 26} fill={grahaColor[g.id]} textAnchor="middle" fontSize="10">{g.rashi_degree.toFixed(1)}°{g.retrograde ? ' ℞' : ''}</text>
          </g>
        );
      })}
      <circle cx={c} cy={c} r="64" className="c-centre" />
      <text x={c} y={c - 14} textAnchor="middle" className="c-lagna-text">LAGNA</text>
      <text x={c} y={c + 8} textAnchor="middle" className="c-title c-mid">{chart.ascendant.rashi}</text>
      <text x={c} y={c + 28} textAnchor="middle" className="c-sub">{chart.ascendant.degree.toFixed(2)}°</text>
    </svg>
  );
}
