import { useEffect, useState } from 'react';
import { getTransits } from '../../api/client';
import type { ChartInput, TransitsResponse } from '../../api/types';
import { grahaEnglish } from '../../astro/format';
import { Card, Chip, Notice, Skeleton } from '../../ds';

const ord = (n: number) => n + (n % 100 >= 11 && n % 100 <= 13 ? 'th' : n % 10 === 1 ? 'st' : n % 10 === 2 ? 'nd' : n % 10 === 3 ? 'rd' : 'th');
const order = ['moon', 'jupiter', 'saturn', 'sun', 'mars', 'mercury', 'venus'];
const weakWords: Record<string, string> = { debilitated: 'in its weakest sign', enemy: 'in an unfriendly sign', combust: 'too close to the Sun' };

/** The planets now, read from the natal Moon with Brihat Samhita 104, in plain words. */
export function TransitNow({ birth }: { birth: ChartInput }) {
  const [data, setData] = useState<TransitsResponse>();
  const [error, setError] = useState('');
  useEffect(() => {
    const ctrl = new AbortController();
    getTransits(birth, 1, ctrl.signal).then(setData).catch(e => { if (e.name !== 'AbortError') setError(e.message); });
    return () => ctrl.abort();
  }, [birth]);
  if (error) return null;
  if (!data) return <Card title="The planets for you now"><Skeleton lines={4} /></Card>;
  const planets = order.map(id => data.planets.find(p => p.graha === id)).filter(p => p !== undefined);
  const saturn = data.sade_sati ? `Sade Sati is running (${['', 'rising', 'peak', 'setting'][data.sade_sati_phase]} phase): Saturn is passing over your Moon sign. Traditionally a time for patience, steady effort and care, not misfortune.`
    : data.kantaka_shani ? 'Saturn is in the 4th house from your Moon (Kantaka Shani): traditionally a time to look after home life and peace of mind.'
      : data.ashtama_shani ? 'Saturn is in the 8th house from your Moon (Ashtama Shani): traditionally a time for caution and patience.' : '';
  return (
    <Card title="The planets for you now" sub="Read from your Moon sign, as the classical books do. The Moon sets the tone of the day, the fast planets the month, the slow ones the year.">
      {saturn && <Notice tone="warning">{saturn}</Notice>}
      <ul className="transit-now">
        {planets.map(p => (
          <li key={p.graha}>
            <div className="transit-now-head">
              <b>{p.graha === 'moon' ? 'Today: the Moon' : grahaEnglish(p.graha)}</b>
              <span className="muted small">in {p.rashi} · your {ord(p.from_moon)} house from the Moon</span>
              {p.favourable !== undefined && <Chip tone={p.favourable ? 'success' : 'warning'}>{p.favourable ? 'Good house' : 'Harder house'}</Chip>}
            </div>
            {p.book && <p>{p.book.plain} <span className="muted small">({p.book.source.replace(/, tr\..*$/, '')} {p.book.ref})</span></p>}
            {(!p.active || p.weakened.length > 0) && (
              <p className="muted small">
                {!p.active && 'Not yet acting: the book says it gives this result in the other half of the sign. '}
                {p.weakened.length > 0 && `Its good results are weaker now: it is ${p.weakened.map(w => weakWords[w] ?? w).join(' and ')}.`}
              </p>
            )}
          </li>
        ))}
      </ul>
      <p className="muted small">See Timing for the coming months.</p>
    </Card>
  );
}
