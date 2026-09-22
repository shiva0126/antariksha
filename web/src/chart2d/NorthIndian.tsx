import { grahaColor } from '../astro/rashi';
import { signIndex, houseNumber } from '../astro/houses';
import { type ChartProps, keyAct, label, useCompact } from './common';

// Houses run counter-clockwise from the top-centre diamond (house 1). For each:
// sign-number position, planet column x, first planet y, and stack direction.
// Geometry: 600-unit square offset by 10; diamonds are houses 1, 4, 7, 10.
const layout: { label: [number, number]; x: number; y: number; dir: 1 | -1 }[] = [
  { label: [310, 48], x: 310, y: 88, dir: 1 },
  { label: [160, 30], x: 160, y: 56, dir: 1 },
  { label: [28, 164], x: 80, y: 130, dir: 1 },
  { label: [160, 196], x: 160, y: 238, dir: 1 },
  { label: [28, 464], x: 80, y: 430, dir: 1 },
  { label: [160, 600], x: 160, y: 578, dir: -1 },
  { label: [310, 348], x: 310, y: 388, dir: 1 },
  { label: [460, 600], x: 460, y: 578, dir: -1 },
  { label: [592, 464], x: 540, y: 430, dir: 1 },
  { label: [460, 196], x: 460, y: 238, dir: 1 },
  { label: [592, 164], x: 540, y: 130, dir: 1 },
  { label: [460, 30], x: 460, y: 56, dir: 1 },
];

export function NorthIndian({ chart, onSelect, selected }: ChartProps) {
  const asc = signIndex(chart.ascendant.longitude);
  const compact = useCompact();
  return (
    <svg className={'chart-svg' + (compact ? ' compact' : '')} viewBox="0 0 620 620" role="img" aria-label="North Indian Kundali with fixed houses">
      <rect x="10" y="10" width="600" height="600" className="c-frame" />
      <path d="M310 10L460 160L310 310L160 160Z" className="c-lagna" />
      <path d="M10 10L610 610M610 10L10 610M310 10L610 310L310 610L10 310Z" className="c-lines" />
      {layout.map((h, i) => {
        const sign = (asc + i) % 12;
        const planets = chart.grahas.filter(g => houseNumber(g.longitude, chart.ascendant.longitude) === i + 1);
        const diamond = i % 3 === 0;
        const step = compact ? Math.min(diamond ? 26 : 22, (diamond ? 104 : 70) / Math.max(planets.length, 1)) : Math.min(diamond ? 20 : 16, (diamond ? 100 : 66) / Math.max(planets.length, 1));
        return (
          <g key={i} data-house={i + 1}>
            <text x={h.label[0]} y={h.label[1]} textAnchor="middle" className={i === 0 ? 'c-signnum c-lagna-text' : 'c-signnum'}>{sign + 1}</text>
            {planets.map((g, j) => (
              <text key={g.id} x={h.x} y={h.y + h.dir * j * step} textAnchor="middle" fill={grahaColor[g.id]}
                className={'c-planet' + (diamond ? '' : ' c-small') + (selected === g.id ? ' is-selected' : '')} role="button" tabIndex={0}
                aria-label={`${g.name} in house ${i + 1}, ${g.rashi} ${g.rashi_degree.toFixed(2)} degrees`}
                onClick={() => onSelect?.(g)} onKeyDown={keyAct(() => onSelect?.(g))}>
                {label(g, compact)}
              </text>
            ))}
          </g>
        );
      })}
    </svg>
  );
}
