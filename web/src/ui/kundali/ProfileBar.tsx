import type { ChartResponse, ReadingResponse } from '../../api/types';
import { grahaEnglish, longDate } from '../../astro/format';
import type { BirthDetails } from './BirthForm';

export function ProfileBar({ birth, chart, reading, onEdit }: { birth: BirthDetails; chart: ChartResponse; reading?: ReadingResponse; onEdit: () => void }) {
  const moon = chart.grahas.find(g => g.id === 'moon');
  const sun = chart.grahas.find(g => g.id === 'sun');
  const dasha = reading?.facts.vimshottari.current;
  const facts: [string, string][] = [
    ['Lagna', `${chart.ascendant.rashi} ${chart.ascendant.degree.toFixed(1)}°`],
    ['Moon sign', moon?.rashi ?? '—'],
    ['Nakshatra', moon ? `${moon.nakshatra} · pada ${moon.nakshatra_pada}` : '—'],
    ['Sun sign', sun?.rashi ?? '—'],
    ['Current dasha', dasha?.maha ? `${grahaEnglish(dasha.maha)} – ${grahaEnglish(dasha.antara ?? '')}` : reading ? '—' : '…'],
  ];
  return (
    <section className="profile card" aria-label="Birth details">
      <div className="profile-id">
        <p className="kicker">Janma kundali</p>
        <h1>{birth.name || 'Your chart'}</h1>
        <p className="muted">{longDate(birth.date)} · {birth.time} · {birth.place} · {birth.tz}</p>
        <button className="ghost" onClick={onEdit}>Edit birth details</button>
      </div>
      <dl className="profile-facts">
        {facts.map(([k, v]) => <div key={k}><dt>{k}</dt><dd>{v}</dd></div>)}
      </dl>
    </section>
  );
}
