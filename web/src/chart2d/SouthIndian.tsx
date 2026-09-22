import { rashis, grahaColor } from '../astro/rashi';
import { signIndex, houseNumber } from '../astro/houses';
import { type ChartProps, keyAct, label, useCompact } from './common';

// Fixed signs clockwise from Meena at the top-left; Mesha is second on top.
const cells = [[1, 0], [2, 0], [3, 0], [3, 1], [3, 2], [3, 3], [2, 3], [1, 3], [0, 3], [0, 2], [0, 1], [0, 0]];
const S = 150, O = 10;

export function SouthIndian({ chart, onSelect, selected, title = 'Rashi · D1' }: ChartProps) {
  const asc = signIndex(chart.ascendant.longitude);
  const compact = useCompact();
  return (
    <svg className={'chart-svg' + (compact ? ' compact' : '')} viewBox="0 0 620 620" role="img" aria-label="South Indian Kundali with fixed zodiac signs">
      <rect x={O} y={O} width={S * 4} height={S * 4} className="c-frame" />
      <rect x={O + S} y={O + S} width={S * 2} height={S * 2} className="c-centre" />
      {cells.map(([col, row], sign) => {
        const x = O + col * S, y = O + row * S;
        const planets = chart.grahas.filter(g => signIndex(g.longitude) === sign);
        const step = Math.min(compact ? 27 : 19, (compact ? 104 : 96) / Math.max(planets.length, 1));
        return (
          <g key={sign} data-sign={sign + 1}>
            <rect x={x} y={y} width={S} height={S} className={asc === sign ? 'c-cell c-lagna' : 'c-cell'} />
            {asc === sign && <path d={`M${x} ${y + 30} L${x + 30} ${y}`} className="c-lagna-mark" />}
            <text x={x + S - 8} y={y + (compact ? 22 : 18)} textAnchor="end" className="c-sign">{compact ? rashis[sign][0].slice(0, 4) : rashis[sign][0]}</text>
            <text x={x + 8} y={y + S - 9} className="c-house">{houseNumber(sign * 30, chart.ascendant.longitude)}</text>
            {asc === sign && !compact && <text x={x + 36} y={y + 18} className="c-lagna-text">Asc {chart.ascendant.degree.toFixed(1)}°</text>}
            {planets.map((g, i) => (
              <text key={g.id} x={x + S / 2} y={y + (compact ? 56 : 46) + i * step} textAnchor="middle" fill={grahaColor[g.id]}
                className={'c-planet' + (selected === g.id ? ' is-selected' : '')} role="button" tabIndex={0}
                aria-label={`${g.name} in ${g.rashi} ${g.rashi_degree.toFixed(2)} degrees`}
                onClick={() => onSelect?.(g)} onKeyDown={keyAct(() => onSelect?.(g))}>
                {label(g, compact)}
              </text>
            ))}
          </g>
        );
      })}
      <text x="310" y="292" textAnchor="middle" className="c-title">{title}</text>
      <text x="310" y="320" textAnchor="middle" className="c-sub">South Indian</text>
      <text x="310" y="346" textAnchor="middle" className="c-lagna-text">Lagna {chart.ascendant.rashi}</text>
    </svg>
  );
}
