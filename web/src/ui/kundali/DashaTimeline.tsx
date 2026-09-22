import type { ChartFacts } from '../../api/types';
import { grahaColor } from '../../astro/rashi';
import { grahaEnglish, longDate } from '../../astro/format';

const years = (from: string, to: string) => (Date.parse(to) - Date.parse(from)) / (365.2425 * 864e5);

export function DashaTimeline({ facts }: { facts: ChartFacts }) {
  const v = facts.vimshottari;
  const today = facts.as_of.slice(0, 10);
  const periods = v.sequence.slice(0, 10);
  const total = years(periods[0].from, periods[periods.length - 1].to);
  return (
    <div className="dasha">
      <div className="card dasha-now">
        <p className="kicker">Running now · {longDate(today)}</p>
        {v.current.maha ? (
          <>
            <h2><span style={{ color: grahaColor[v.current.maha] }}>{grahaEnglish(v.current.maha)}</span> mahadasha · <span style={{ color: grahaColor[v.current.antara ?? ''] }}>{grahaEnglish(v.current.antara ?? '')}</span> antardasha</h2>
            <p className="muted">Antardasha {longDate(v.current.from)} → {longDate(v.current.to)}{v.upcoming.lord && <> · next mahadasha {grahaEnglish(v.upcoming.lord)} from {longDate(v.upcoming.from)}</>}</p>
          </>
        ) : <p>No running period for this date.</p>}
        <p className="muted small">Birth balance: {grahaEnglish(v.birth_balance.lord)} mahadasha with {v.birth_balance.years_remaining.toFixed(2)} years remaining at birth (from the Moon's position in its nakshatra).</p>
      </div>
      <div className="card">
        <h3>Vimshottari mahadashas</h3>
        <div className="dasha-bar" aria-hidden>
          {periods.map(p => (
            <span key={p.from} style={{ flexGrow: years(p.from, p.to) / total, background: grahaColor[p.lord] }} className={p.from <= today && today < p.to ? 'is-current' : ''} title={grahaEnglish(p.lord)} />
          ))}
        </div>
        <ol className="dasha-list">
          {periods.map(p => {
            const current = p.from <= today && today < p.to, past = p.to <= today;
            return (
              <li key={p.from} className={current ? 'is-current' : past ? 'is-past' : ''}>
                <i style={{ background: grahaColor[p.lord] }} />
                <b>{grahaEnglish(p.lord)}</b>
                <span>{longDate(p.from)} – {longDate(p.to)}</span>
                <em>{years(p.from, p.to).toFixed(1)} y{current ? ' · now' : ''}</em>
              </li>
            );
          })}
        </ol>
      </div>
      <div className="grid-2">
        <SubList title={`Antardashas of ${grahaEnglish(v.current.maha ?? '')}`} periods={v.antaras ?? []} today={today} />
        <SubList title={`Pratyantardashas of ${grahaEnglish(v.current.antara ?? '')}`} periods={v.pratyantaras ?? []} today={today} />
      </div>
      {facts.yogini?.sequence?.length > 0 && (
        <div className="card">
          <h3>Yogini dasha</h3>
          <p className="muted small">The 36-year cycle of eight yoginis, from the Moon's nakshatra. Running now: <b>{facts.yogini.current.yogini}</b> ({grahaEnglish(facts.yogini.current.lord)}) until {longDate(facts.yogini.current.to)}.</p>
          <ol className="dasha-list">
            {facts.yogini.sequence.slice(0, 10).map(p => {
              const current = p.from <= today && today < p.to, past = p.to <= today;
              return <li key={p.from} className={current ? 'is-current' : past ? 'is-past' : ''}><i style={{ background: grahaColor[p.lord] }} /><b>{p.yogini}</b><span>{longDate(p.from)} – {longDate(p.to)}</span><em>{grahaEnglish(p.lord)}{current ? ' · now' : ''}</em></li>;
            })}
          </ol>
        </div>
      )}
    </div>
  );
}

function SubList({ title, periods, today }: { title: string; periods: { lord: string; from: string; to: string }[]; today: string }) {
  if (!periods.length) return null;
  return (
    <div className="card">
      <h3>{title}</h3>
      <ol className="dasha-list compact">
        {periods.map(p => {
          const current = p.from <= today && today < p.to, past = p.to <= today;
          return <li key={p.from} className={current ? 'is-current' : past ? 'is-past' : ''}><i style={{ background: grahaColor[p.lord] }} /><b>{grahaEnglish(p.lord)}</b><span>{longDate(p.from)} – {longDate(p.to)}</span>{current && <em>now</em>}</li>;
        })}
      </ol>
    </div>
  );
}
