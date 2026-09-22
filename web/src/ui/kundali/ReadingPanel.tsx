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
        <p className="muted small">{data.model === 'grounded-corpus' ? 'Composed from the chart facts and the Antariksha corpus.' : `Written by ${data.model} from the chart facts and corpus.`} {(data.grounding ?? []).length} passages consulted{classical.length ? `, ${classical.length} from classical texts` : ''}.</p>
      </div>
      <article className="card"><h3>Lagna and Moon</h3><p>{r.lagna_and_moon}</p></article>
      <div className="grid-2">
        {themes.map(([k, v]) => <article className="card" key={k}><h3>{k}</h3><p>{v || '—'}</p></article>)}
      </div>
      <article className="card">
        <h3>Yogas · {r.yogas.length}</h3>
        {r.yogas.length === 0 && <p className="muted">No yoga from the engine's catalogue is present.</p>}
        {r.yogas.map(y => (
          <div className="row" key={y.name}>
            <h4>{y.name} <span className={'badge badge-' + y.strength}>{y.strength}</span></h4>
            <p>{y.meaning}</p>
            {y.effect && <p className="muted small">{y.effect}</p>}
          </div>
        ))}
      </article>
      <article className="card">
        <h3>Planets</h3>
        {r.grahas.map(g => (
          <div className="row" key={g.graha}>
            <h4>{g.graha}</h4>
            <p className="muted small">{g.placement}</p>
            <p>{g.meaning}</p>
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
