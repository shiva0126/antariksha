import { useEffect, useState } from 'react';
import { postProfileChat } from '../../api/client';
import type { MatchResponse } from '../../api/types';
import { useNames } from '../../i18n';
import { Explanation } from '../match/Explanation';
import { MatchChat } from '../match/MatchChat';
import { memberAPI } from '../community/shared';
import { GRAHA } from './options';
import { displayName, type Candidate, type Horoscope } from './types';

type CompareResponse = MatchResponse & { horoscope: Horoscope; groom_has_place: boolean; bride_has_place: boolean };

/** Full kundali comparison with a member, computed on the server from both
 *  members' stored birth details (never shown), plus Ask Astrisk. */
export function Compare({ peer }: { peer: Candidate }) {
  const n = useNames();
  const [data, setData] = useState<CompareResponse>(), [error, setError] = useState('');
  useEffect(() => { memberAPI<CompareResponse>('/api/matrimony/compare/' + peer.id).then(setData).catch(e => setError(e.message)); }, [peer.id]);
  if (error) return <p role="alert" className="form-error">{error}</p>;
  if (!data) return <p className="muted">Comparing both kundalis…</p>;
  const m = data.match, h = data.horoscope, them = displayName(peer);
  const groom = h.viewer_is_groom_side ? 'You' : them, bride = h.viewer_is_groom_side ? them : 'You';
  return (
    <div className="mat-compare">
      <div className="match-score">
        <div className="score-ring" style={{ '--pct': Math.round(m.total / m.max * 100) } as React.CSSProperties} role="img" aria-label={`${m.total} of ${m.max} gunas`}><b>{m.total}</b><span>of {m.max}</span></div>
        <div>
          <p><b>{groom}</b> (groom side): {m.boy_moon}<br /><b>{bride}</b> (bride side): {m.girl_moon}</p>
          {m.doshas.length ? <ul className="tag-list">{m.doshas.map(d => <li key={d} className="tag tag-caution">{d}</li>)}</ul> : <p className="good-text">No koota dosha.</p>}
          {m.exceptions.map(x => <p key={x} className="muted small">✓ {x}</p>)}
          <p className="small">Mangal dosha: {groom} {data.groom_has_place ? (m.boy_mangal_dosha ? 'present' : 'not present') : 'needs birthplace'}; {bride} {data.bride_has_place ? (m.girl_mangal_dosha ? 'present' : 'not present') : 'needs birthplace'}.</p>
          {h.their_root_graha && <p className="small">Numerology: your root number {h.your_root_number} ({GRAHA[h.your_root_graha]}) and theirs {h.their_root_number} ({GRAHA[h.their_root_graha]}) are {h.number_relation}.</p>}
        </div>
      </div>
      <div className="table-scroll">
        <table className="planet-table koota-table">
          <thead><tr><th>Koota</th><th>{groom}</th><th>{bride}</th><th>Score</th></tr></thead>
          <tbody>{m.kootas.map(k => <tr key={k.name}><td><b>{k.name}</b></td><td>{n(k.boy)}</td><td>{n(k.girl)}</td><td className="num"><span className={'koota-score ' + (k.score === k.max ? 'full' : k.score === 0 ? 'zero' : '')}>{k.score} / {k.max}</span></td></tr>)}</tbody>
        </table>
      </div>
      {data.explanation && <details><summary>What each factor means</summary><Explanation report={data.explanation} /></details>}
      <MatchChat ask={(q, hist, lang) => postProfileChat(peer.id, q, hist, lang)} intro={`Ask about your match with ${them}. Answers use both kundalis and both root numbers; birth details stay private. Nothing here is saved.`} />
    </div>
  );
}
