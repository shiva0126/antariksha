import { useEffect, useState } from 'react';
import { useT } from '../../i18n';
import { Dialog } from '../common/Dialog';
import { Avatar, memberAPI } from '../community/shared';
import { PhotoGallery } from '../community/Biodata';
import { Compare } from './Compare';
import { ProfileExplanation } from '../match/Explanation';
import { GRAHA, heightLabel, label, OPTIONS } from './options';
import { displayName, type Candidate, type DiscoverResponse } from './types';

const FILTERS: [string, string][] = [['religion', 'Religion'], ['mother_tongue', 'Mother tongue'], ['diet', 'Diet'], ['marital_status', 'Marital status'], ['education_level', 'Education']];

/** A profile card: what a family looks at first, with the reasons it was suggested. */
export function ProfileCard({ c, onChanged, onCompare }: { c: Candidate; onChanged: (c: Candidate) => void; onCompare: (c: Candidate) => void }) {
  const t = useT();
  const [interestOpen, setInterestOpen] = useState(false), [reportOpen, setReportOpen] = useState(false), [explain, setExplain] = useState(false);
  const [note, setNote] = useState(''), [reason, setReason] = useState(''), [error, setError] = useState(''), [busy, setBusy] = useState(false), [open, setOpen] = useState(false);
  const d = c.details, h = c.horoscope;
  async function act(fn: () => Promise<void>) { setBusy(true); setError(''); try { await fn(); } catch (e) { setError(e instanceof Error ? e.message : 'Request failed'); } finally { setBusy(false); } }
  const save = (kind: 'saved' | 'skipped' | 'clear') => act(async () => { await memberAPI('/api/matrimony/saved', 'POST', { target: c.id, kind }); onChanged({ ...c, saved: kind === 'clear' ? '' : kind }); });
  const facts = [d.city, heightLabel(d.height_cm), label('religion', d.religion), label('mother_tongue', d.mother_tongue), label('diet', d.diet)].filter(Boolean);
  const work = [d.occupation || label('occupation_category', d.occupation_category), d.education || label('education_level', d.education_level)].filter(Boolean);
  return (
    <article className="card mat-card" aria-label={displayName(c)}>
      <div className="mat-card-media">
        {c.photos.length ? <PhotoGallery photos={c.photos.slice(0, 1)} /> : <div className="mat-photo-empty"><Avatar kind={c.avatar} accent={c.accent} /></div>}
        {h && <div className="guna-chip" title={`${h.guna} of ${h.max} gunas`}><b>{h.guna}</b><span>/36</span></div>}
      </div>
      <div className="mat-card-body">
        <h3>{displayName(c)} · {c.age}{c.verified && <span className="verified" title="A moderator compared a selfie with profile photos. This does not verify identity documents, income, occupation or character.">✓ Photo reviewed</span>}</h3>
        <p className="mat-facts">{facts.join(' · ')}</p>
        {work.length > 0 && <p className="mat-facts">{work.join(' · ')}</p>}
        {d.introduction && <p className="mat-intro">{d.introduction}</p>}
        {c.reasons.length > 0 && <ul className="reason-chips" aria-label={t('Why suggested')}>{c.reasons.map(r => <li key={r} className={r.startsWith('Discuss') ? 'discuss' : ''}>{r}</li>)}</ul>}
        {h ? (
          <p className="mat-horo small">Moon {h.their_moon} · Mangal dosha: {h.their_mangal === 'unknown' ? 'needs birthplace' : h.their_mangal}{h.their_root_graha && <> · root number {h.their_root_number} ({GRAHA[h.their_root_graha]}), {h.number_relation} with yours</>}</p>
        ) : c.horoscope_note && <p className="muted small">{c.horoscope_note}</p>}
        <details open={open} onToggle={e => setOpen((e.target as HTMLDetailsElement).open)}>
          <summary>Full biodata</summary>
          <dl className="kv">
            {([['Marital status', label('marital_status', d.marital_status)], ['Family type', label('family_type', d.family_type)], ['Income', label('income_band', d.income_band)], ['Marriage timeline', label('timeline', d.timeline)], ['Relocation', label('relocation', d.relocation)], ['Children', label('children', d.children)], ['Languages', d.languages], ['Community', d.community], ['Hobbies', d.hobbies], ['Values', d.values], ['About the family', d.family_about]] as [string, string | undefined][]).filter(([, v]) => v).map(([k, v]) => <div key={k}><dt>{t(k)}</dt><dd>{v}</dd></div>)}
          </dl>
          {c.photos.length > 1 && <PhotoGallery photos={c.photos.slice(1)} />}
          {(d.social_links || []).map(link => <p key={link} className="small"><a href={link} target="_blank" rel="noopener noreferrer">Social profile (self-declared)</a></p>)}
        </details>
        <div className="mat-actions">
          {c.interest === '' && <button className="primary" disabled={busy} onClick={() => setInterestOpen(true)}>{t('Express interest')}</button>}
          {c.interest === 'sent' && <span className="tag">Interest sent</span>}
          {c.interest === 'received' && <span className="tag">Interested in you — see Interests</span>}
          {c.interest === 'accepted' && <span className="tag tag-good">Connected</span>}
          {c.saved === 'saved' ? <button className="ghost" disabled={busy} onClick={() => void save('clear')}>★ {t('Saved')}</button> : <button className="ghost" disabled={busy} onClick={() => void save('saved')}>☆ {t('Save')}</button>}
          {c.saved === 'skipped' ? <button className="ghost" disabled={busy} onClick={() => void save('clear')}>Show again</button> : <button className="ghost" disabled={busy} onClick={() => void save('skipped')}>{t('Not now')}</button>}
          {h && <button className="ghost" onClick={() => onCompare(c)}>{t('Compare horoscopes')}</button>}
          <button className="ghost" onClick={() => setExplain(true)}>Compare our answers</button>
          <button className="ghost danger-text" onClick={() => setReportOpen(true)}>{t('Report profile')}</button>
        </div>
        {error && <p role="alert" className="form-error">{error}</p>}
      </div>
      <Dialog open={interestOpen} onClose={() => setInterestOpen(false)} title={`${t('Express interest')}: ${displayName(c)}`}>
        <form onSubmit={e => { e.preventDefault(); void act(async () => { await memberAPI('/api/matrimony/interests', 'POST', { target: c.id, action: 'send', note }); setInterestOpen(false); onChanged({ ...c, interest: 'sent' }); }); }}>
          <label className="field"><span>A short note (optional)</span><textarea maxLength={300} rows={4} value={note} onChange={e => setNote(e.target.value)} placeholder={c.shared_interests.length ? `For example: we both enjoy ${c.shared_interests[0]}…` : 'Say why you would like to talk'} /></label>
          <p className="muted small">{300 - note.length} characters left. Your note is shown only to {displayName(c)}. Conversations open when they accept.</p>
          {error && <p role="alert" className="form-error">{error}</p>}
          <button className="primary" disabled={busy}>{t('Send')}</button>
        </form>
      </Dialog>
      <Dialog open={explain} onClose={() => setExplain(false)} title={`Compare our answers: ${displayName(c)}`}>
        <ProfileExplanation peer={c.id} />
      </Dialog>
      <Dialog open={reportOpen} onClose={() => setReportOpen(false)} title={`${t('Report profile')}: ${displayName(c)}`}>
        <form onSubmit={e => { e.preventDefault(); void act(async () => { await memberAPI('/api/matrimony/report/' + c.id, 'POST', { reason }); setReportOpen(false); setReason(''); }); }}>
          <label className="field"><span>What is the concern?</span>
            <select required value={reason.split(':')[0]} onChange={e => setReason(e.target.value + ': ')}>
              <option value="">Choose a reason</option>
              {['Fake or misleading profile', 'Asking for money', 'Harassment or abuse', 'Underage', 'Inappropriate photos', 'Other'].map(r => <option key={r}>{r}</option>)}
            </select></label>
          <label className="field"><span>Details for the moderator</span><textarea required minLength={3} maxLength={500} rows={3} value={reason} onChange={e => setReason(e.target.value)} /></label>
          <p className="muted small">Moderators review every report. If anyone asks you for money or you feel unsafe, stop replying and block them. In an emergency call 112.</p>
          <button className="primary" disabled={busy}>Send report</button>
        </form>
      </Dialog>
    </article>
  );
}

export function Discover({ saved: savedSearch, horoscopeOn }: { saved: Record<string, string>; horoscopeOn: boolean }) {
  const t = useT();
  const [filters, setFilters] = useState<Record<string, string>>(savedSearch), [applied, setApplied] = useState<Record<string, string>>(savedSearch);
  const [page, setPage] = useState(0), [data, setData] = useState<DiscoverResponse>(), [error, setError] = useState(''), [msg, setMsg] = useState(''), [compare, setCompare] = useState<Candidate>();
  const [showFilters, setShowFilters] = useState(false);
  useEffect(() => {
    const q = new URLSearchParams(Object.entries({ ...applied, page: String(page) }).filter(([, v]) => v !== '' && v !== undefined));
    setError('');
    memberAPI<DiscoverResponse>('/api/matrimony/discover?' + q).then(setData).catch(e => setError(e.message));
  }, [applied, page]);
  const set = (k: string, v: string) => setFilters(f => ({ ...f, [k]: v }));
  const apply = () => { setApplied(filters); setPage(0); setShowFilters(false); };
  const replace = (c: Candidate) => setData(d => d && { ...d, items: d.items.map(x => x.id === c.id ? c : x) });
  return (
    <div className="mat-discover">
      <aside className={'card mat-filters' + (showFilters ? ' open' : '')} aria-label={t('Filters')}>
        <div className="mat-filters-head"><h3>{t('Filters')}</h3><button className="ghost small mat-filters-close" onClick={() => setShowFilters(false)}>✕</button></div>
        <label className="field"><span>Looking for</span><select value={filters.kind ?? ''} onChange={e => set('kind', e.target.value)}><option value="">Anyone</option><option value="bride">Bride</option><option value="groom">Groom</option></select></label>
        <div className="field-row"><label className="field"><span>{t('Age')} from</span><input type="number" min={18} max={100} value={filters.age_min ?? ''} onChange={e => set('age_min', e.target.value)} /></label>
          <label className="field"><span>to</span><input type="number" min={18} max={100} value={filters.age_max ?? ''} onChange={e => set('age_max', e.target.value)} /></label></div>
        <label className="field"><span>{t('City')}</span><input value={filters.city ?? ''} onChange={e => set('city', e.target.value)} placeholder="Any city" /></label>
        {FILTERS.map(([k, l]) => <label className="field" key={k}><span>{t(l)}</span><select value={filters[k] ?? ''} onChange={e => set(k, e.target.value)}><option value="">Any</option>{OPTIONS[k].map(([v, lab]) => <option key={v} value={v}>{lab}</option>)}</select></label>)}
        <fieldset className="mat-horo-filters" disabled={!horoscopeOn}>
          <legend>Horoscope {horoscopeOn ? '' : '(turn on horoscope matching in your profile)'}</legend>
          <label className="field"><span>Minimum {t('Guna score')}</span><select value={filters.min_guna ?? ''} onChange={e => set('min_guna', e.target.value)}><option value="">Any</option>{[12, 18, 21, 24, 28].map(n => <option key={n} value={n}>{n}+ of 36</option>)}</select></label>
          <label className="field"><span>Mangal dosha</span><select value={filters.mangal ?? ''} onChange={e => set('mangal', e.target.value)}><option value="">Any</option><option value="no">Not present</option><option value="yes">Present</option></select></label>
        </fieldset>
        <label className="consent"><input type="checkbox" checked={filters.verified === '1'} onChange={e => set('verified', e.target.checked ? '1' : '')} /> <span>Photo verified only</span></label>
        <div className="mat-filter-actions">
          <button className="primary" onClick={apply}>{t('Apply filters')}</button>
          <button className="ghost" onClick={() => { setFilters({}); setApplied({}); setPage(0); }}>{t('Clear')}</button>
          <button className="ghost" onClick={() => { memberAPI('/api/matrimony/search', 'PUT', filters).then(() => setMsg('Search saved. It opens with these filters next time.')).catch(e => setMsg(e.message)); }}>{t('Save this search')}</button>
        </div>
        {msg && <p role="status" className="small">{msg}</p>}
        <p className="muted small">A guna score is a traditional table, not a prediction. Filtering by it can hide people you would get on with.</p>
      </aside>
      <section className="mat-results" aria-label="Profiles">
        <div className="mat-results-head">
          <p><b>{data?.total ?? '…'}</b> profiles</p>
          <button className="ghost mat-filters-toggle" onClick={() => setShowFilters(true)}>☰ {t('Filters')}</button>
          <label className="field compact"><span className="sr-only">{t('Sort by')}</span>
            <select aria-label={t('Sort by')} value={applied.sort ?? ''} onChange={e => { setApplied(a => ({ ...a, sort: e.target.value })); setFilters(f => ({ ...f, sort: e.target.value })); setPage(0); }}>
              <option value="">{t('Best match')}</option><option value="guna">{t('Guna score')}</option><option value="recent">{t('Newest')}</option></select></label>
          <label className="consent"><input type="checkbox" checked={applied.saved === '1'} onChange={e => { setApplied(a => ({ ...a, saved: e.target.checked ? '1' : '' })); setPage(0); }} /> <span>{t('Shortlist')} only</span></label>
        </div>
        {error && <p role="alert" className="form-error">{error}</p>}
        {data?.needs_profile && <div className="card"><p>Create your matrimony profile first, then discovery shows people whose age preferences include you and yours include them.</p></div>}
        {data && !data.needs_profile && data.total === 0 && <div className="card"><p>No profiles match these filters yet. Try fewer filters, or check back as members join.</p></div>}
        <div className="mat-grid">{data?.items.map(c => <ProfileCard key={c.id} c={c} onChanged={replace} onCompare={setCompare} />)}</div>
        {data && data.pages > 1 && (
          <nav className="pager" aria-label="Pages">
            <button className="ghost" disabled={page === 0} onClick={() => setPage(p => p - 1)}>← {t('Previous')}</button>
            <span>{page + 1} / {data.pages}</span>
            <button className="ghost" disabled={page + 1 >= data.pages} onClick={() => setPage(p => p + 1)}>{t('Next')} →</button>
          </nav>
        )}
      </section>
      <Dialog open={!!compare} onClose={() => setCompare(undefined)} title={compare ? `${t('Compare horoscopes')}: ${displayName(compare)}` : ''}>
        {compare && <Compare peer={compare} />}
      </Dialog>
    </div>
  );
}
