import { useMemo, useState } from 'react';
import { postMatch, postMatchChat } from '../../api/client';
import { Explanation } from './Explanation';
import { MatchChat } from './MatchChat';
import type { MatchResponse } from '../../api/types';
import { useNames, useT } from '../../i18n';
import { Button, Checkbox, PageHeader } from '../../ds';
import { loadProfiles } from '../common/profiles';
import { BirthFields, draftFrom, resolveDraft, type BirthDetails, type BirthDraft } from '../kundali/BirthForm';

const ord = (n: number) => (n === 1 ? 'st' : n === 2 ? 'nd' : n === 3 ? 'rd' : 'th');

function ProfilePicker({ onPick }: { onPick: (d: BirthDraft) => void }) {
  const { profiles } = useMemo(loadProfiles, []);
  if (!profiles.length) return null;
  return (
    <label className="field compact"><span>Fill from a saved profile</span>
      <select defaultValue="" onChange={e => { const p = profiles.find(x => x.id === e.target.value); if (p) onPick(draftFrom(p)); }}>
        <option value="" disabled>Choose…</option>
        {profiles.map(p => <option key={p.id} value={p.id}>{p.name || `Chart of ${p.date}`}</option>)}
      </select>
    </label>
  );
}

export function MatchPage() {
  const t = useT(), n = useNames();
  const [boy, setBoy] = useState<BirthDraft>(() => ({ ...draftFrom(), date: '1994-11-02', time: '06:40' }));
  const [girl, setGirl] = useState<BirthDraft>(() => draftFrom());
  const [result, setResult] = useState<MatchResponse>();
  const [pair, setPair] = useState<{ boy: BirthDetails; girl: BirthDetails }>();
  const [useAI,setUseAI]=useState(false);
  const [error, setError] = useState(''), [busy, setBusy] = useState(false);

  function changeBoy(draft: BirthDraft) { setBoy(draft); setResult(undefined); setError(''); }
  function changeGirl(draft: BirthDraft) { setGirl(draft); setResult(undefined); setError(''); }

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    const b = resolveDraft(boy), g = resolveDraft(girl);
    if (typeof b === 'string') return setError(`${t('Groom')}: ${b}`);
    if (typeof g === 'string') return setError(`${t('Bride')}: ${g}`);
    setError(''); setBusy(true);
    setResult(undefined);
    try { setResult(await postMatch(b, g, useAI)); setPair({ boy: b, girl: g }); } catch (err) { setError(err instanceof Error ? err.message : 'Matching failed'); } finally { setBusy(false); }
  }

  const m = result?.match;
  const pct = m ? Math.round((m.total / m.max) * 100) : 0;
  return (
    <div className="page match">
      <PageHeader kicker="Ashtakoota Guna Milan" title={t('Matching')} compactOnMobile
        description="Compare the traditional 36-point Moon-based tables and the Mars rule, with plain-language explanations and Ask Astrisk. Scores are not a prediction of relationship success." />
      <form className="match-form" onSubmit={submit} noValidate>
        <fieldset disabled={busy} className="contents">
        <div className="card">
          <div className="match-side-head"><h3>{t('Groom')}</h3><ProfilePicker onPick={changeBoy} /></div>
          <BirthFields draft={boy} onChange={changeBoy} />
        </div>
        <div className="card">
          <div className="match-side-head"><h3>{t('Bride')}</h3><ProfilePicker onPick={changeGirl} /></div>
          <BirthFields draft={girl} onChange={changeGirl} />
        </div>
        {error && <p role="alert" className="form-error field-wide">{error}</p>}
        <div className="field-wide"><Checkbox label="Use AI for a more detailed explanation" checked={useAI} onChange={e=>{setUseAI(e.target.checked);setResult(undefined);}}/></div>
        <p className="small muted field-wide">AI receives only calculated factor scores and flags—not names or birth details. Use birth information shared with permission. Results are not saved. A complete fact-based guide is available without AI.</p>
        <Button type="submit" variant="primary" block busy={busy} className="field-wide">{t('Match charts')}</Button>
        </fieldset>
      </form>

      {m && result && (
        <section className="match-result" aria-label="Matching result">
          {result.explanation&&<Explanation report={result.explanation}/>}
          <div className="card match-score">
            <div className="score-ring" style={{ '--pct': pct } as React.CSSProperties} role="img" aria-label={`${m.total} of ${m.max} gunas`}>
              <b>{m.total}</b><span>of {m.max}</span>
            </div>
            <div>
              <h2>{m.verdict}</h2>
              <p className="muted">{boy.name || t('Groom')}: {m.boy_moon} · {girl.name || t('Bride')}: {m.girl_moon}</p>
              {m.doshas.length > 0 ? <ul className="tag-list">{m.doshas.map(d => <li key={d} className="tag tag-caution">{d}</li>)}</ul> : <p className="good-text">No koota dosha.</p>}
              {m.exceptions.map(x => <p key={x} className="muted small">✓ {x}</p>)}
            </div>
          </div>
          <div className="card table-card">
            <div className="table-scroll">
              <table className="planet-table koota-table">
                <thead><tr><th>Koota</th><th>{t('Groom')}</th><th>{t('Bride')}</th><th>Score</th><th>Traditional rule and limits</th></tr></thead>
                <tbody>{m.kootas.map(k => (
                  <tr key={k.name}><td><b>{k.name}</b></td><td>{n(k.boy)}</td><td>{n(k.girl)}</td>
                    <td className="num"><span className={'koota-score ' + (k.score === k.max ? 'full' : k.score === 0 ? 'zero' : '')}>{k.score} / {k.max}</span></td>
                    <td className="wrap">{k.description}</td></tr>
                ))}</tbody>
              </table>
            </div>
          </div>
          <div className="grid-2">
            <article className="card"><h3>Mangal dosha</h3>
              <dl className="kv">
                <div><dt>{t('Groom')}</dt><dd>{m.boy_mangal_dosha ? `Present: Mars in the ${result.boy.mars_house}${ord(result.boy.mars_house)} house` : `Not present: Mars in the ${result.boy.mars_house}${ord(result.boy.mars_house)} house`}</dd></div>
                <div><dt>{t('Bride')}</dt><dd>{m.girl_mangal_dosha ? `Present: Mars in the ${result.girl.mars_house}${ord(result.girl.mars_house)} house` : `Not present: Mars in the ${result.girl.mars_house}${ord(result.girl.mars_house)} house`}</dd></div>
              </dl>
              <p className="muted small">{m.mangal_note}</p>
            </article>
            <article className="card"><h3>Charts</h3>
              <dl className="kv">
                <div><dt>{t('Groom')}</dt><dd>{n(result.boy.lagna)} lagna · {n(result.boy.moon_sign)} Moon · {n(result.boy.nakshatra)} {result.boy.pada}</dd></div>
                <div><dt>{t('Bride')}</dt><dd>{n(result.girl.lagna)} lagna · {n(result.girl.moon_sign)} Moon · {n(result.girl.nakshatra)} {result.girl.pada}</dd></div>
              </dl>
              <p className="muted small">Guna Milan is one traditional input. Families usually also compare the full charts (7th house, Venus, Navamsha and dashas). This tool gives reflection, not a verdict.</p>
            </article>
          </div>
          {pair && <MatchChat key={JSON.stringify(pair)} ask={(q, h, lang) => postMatchChat(pair.boy, pair.girl, q, h, lang)} />}
        </section>
      )}
    </div>
  );
}
