import { useMemo, useState } from 'react';
import { postMatch } from '../../api/client';
import type { MatchResponse } from '../../api/types';
import { useNames, useT } from '../../i18n';
import { loadProfiles } from '../common/profiles';
import { BirthFields, draftFrom, resolveDraft, type BirthDraft } from '../kundali/BirthForm';

function ProfilePicker({ onPick }: { onPick: (d: BirthDraft) => void }) {
  const { profiles } = useMemo(loadProfiles, []);
  if (!profiles.length) return null;
  return (
    <label className="field compact"><span>Fill from a saved profile</span>
      <select defaultValue="" onChange={e => { const p = profiles.find(x => x.id === e.target.value); if (p) onPick(draftFrom(p)); }}>
        <option value="" disabled>Choose…</option>
        {profiles.map(p => <option key={p.id} value={p.id}>{p.name || p.date}</option>)}
      </select>
    </label>
  );
}

export function MatchPage() {
  const t = useT(), n = useNames();
  const [boy, setBoy] = useState<BirthDraft>(() => ({ ...draftFrom(), date: '1994-11-02', time: '06:40' }));
  const [girl, setGirl] = useState<BirthDraft>(() => draftFrom());
  const [result, setResult] = useState<MatchResponse>();
  const [error, setError] = useState(''), [busy, setBusy] = useState(false);

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    const b = resolveDraft(boy), g = resolveDraft(girl);
    if (typeof b === 'string') return setError(`${t('Groom')}: ${b}`);
    if (typeof g === 'string') return setError(`${t('Bride')}: ${g}`);
    setError(''); setBusy(true);
    try { setResult(await postMatch(b, g)); } catch (err) { setError(err instanceof Error ? err.message : 'Matching failed'); } finally { setBusy(false); }
  }

  const m = result?.match;
  const pct = m ? Math.round((m.total / m.max) * 100) : 0;
  return (
    <div className="page match">
      <header className="page-head">
        <div>
          <p className="kicker">Ashtakoota Guna Milan</p>
          <h1>{t('Matching')}</h1>
          <p className="muted">The traditional 36-point compatibility check from both Moons, with Nadi, Bhakoot and Gana doshas, their classical cancellations, and Mangal dosha in both charts. Free.</p>
        </div>
      </header>
      <form className="match-form" onSubmit={submit} noValidate>
        <div className="card">
          <div className="match-side-head"><h3>{t('Groom')}</h3><ProfilePicker onPick={setBoy} /></div>
          <BirthFields draft={boy} onChange={setBoy} />
        </div>
        <div className="card">
          <div className="match-side-head"><h3>{t('Bride')}</h3><ProfilePicker onPick={setGirl} /></div>
          <BirthFields draft={girl} onChange={setGirl} />
        </div>
        {error && <p role="alert" className="form-error field-wide">{error}</p>}
        <button className="primary field-wide" disabled={busy}>{busy ? 'Matching…' : t('Match charts')}</button>
      </form>

      {m && result && (
        <section className="match-result" aria-label="Matching result">
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
                <thead><tr><th>Koota</th><th>{t('Groom')}</th><th>{t('Bride')}</th><th>Score</th><th>What it measures</th></tr></thead>
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
                <div><dt>{t('Groom')}</dt><dd>{m.boy_mangal_dosha ? `Present (Mars in house ${result.boy.mars_house})` : `Not present (Mars in house ${result.boy.mars_house})`}</dd></div>
                <div><dt>{t('Bride')}</dt><dd>{m.girl_mangal_dosha ? `Present (Mars in house ${result.girl.mars_house})` : `Not present (Mars in house ${result.girl.mars_house})`}</dd></div>
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
        </section>
      )}
    </div>
  );
}
