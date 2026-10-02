import type { ReadingResponse } from '../../api/types';

export function ReadingPanel({ data }: { data: ReadingResponse }) {
  const r = data.reading;
  const themes: [string, string][] = [['Career', r.themes.career], ['Relationships', r.themes.relationships], ['Strengths', r.themes.strengths], ['Growth areas', r.themes.growth_areas]];
  const classical = (data.grounding ?? []).filter(g => g.source.includes('[public_domain]'));
  return (
    <div className="reading">
      <div className="card reading-lead">
        <p className="kicker">Summary</p>
        <p className="lead">{r.summary}</p>
        <p className="muted small">{data.model === 'grounded-corpus' ? 'Composed from chart facts and Astrisk’s original interpretation guides.' : `AI-assisted wording by ${data.model}, using chart facts and retrieved guides.`} {(data.grounding ?? []).length} entries retrieved{classical.length ? `, including ${classical.length} historical reference records` : ''}. Historical citations are provenance, not proof that a prediction is true.</p>
        <details><summary>Sources and interpretation limits</summary><p>Original guides explain symbolic themes in everyday language. Historical wording is not quoted in this reading. Additional books, including Lal Kitab, are not automatically included. No model has been trained on these entries.</p><ul>{(data.grounding??[]).map((g,i)=><li key={g.key+'-'+i}><strong>{g.title||g.key}</strong> — {g.source}{g.ref?` · ${g.ref}`:''}<br/><span className="small muted">Matched fact: {g.doc_type}:{g.key}</span></li>)}</ul></details>
      </div>
      <article className="card"><h3>Your approach to life and emotions</h3><p>{r.lagna_and_moon}</p></article>
      <div className="grid-2">
        {themes.map(([k, v]) => <article className="card" key={k}><h3>{k}</h3><p>{v || '—'}</p></article>)}
      </div>
      <article className="card">
        <h3>Patterns in your chart · {r.yogas.length}</h3>
        <p className="muted small">These combinations are traditionally called yogas. Their strength labels describe the engine's rule—not your worth or the certainty of an outcome.</p>
        {r.yogas.length === 0 && <p className="muted">No yoga from the engine's catalogue is present.</p>}
        {r.yogas.map((y,i) => (
          <div className="row" key={y.name+'-'+i}>
            <h4>{y.name} <span className={'badge badge-' + y.strength}>{y.strength}</span></h4>
            <p>{y.meaning}</p>
            {y.effect && <details><summary>Chart details</summary><p className="muted small">{y.effect}</p></details>}
          </div>
        ))}
      </article>
      <article className="card">
        <h3>Planets</h3>
        {r.grahas.map(g => (
          <div className="row" key={g.graha}>
            <h4>{g.graha}</h4>
            <p>{g.meaning}</p>
            <details><summary>Position details</summary><p className="muted small">{g.placement}</p></details>
          </div>
        ))}
      </article>
      <div className="grid-2">
        <article className="card"><h3>Current period</h3><p>{r.dashas.current}</p></article>
        <article className="card"><h3>Next period</h3><p>{r.dashas.upcoming || '—'}</p></article>
      </div>
      <p className="disclaimer">{r.disclaimer}</p>
    </div>
  );
}
