import { useEffect, useState } from 'react';
import { useT } from '../../i18n';
import { PlaceSearch } from '../common/PlaceSearch';
import type { Place } from '../common/places';
import { PhotoEditor, cleanBiodata, emptyBiodata, type MatrimonyPhoto } from '../community/Biodata';
import { memberAPI } from '../community/shared';
import { OPTIONS } from './options';
import { disablePush, enablePush, pushEnabled, pushSupported } from './push';
import type { Details, MyProfile } from './types';

const IMPORTANT: (keyof Details)[] = ['display_name', 'profile_kind', 'introduction', 'city', 'religion', 'mother_tongue', 'diet', 'height_cm', 'marital_status', 'education_level', 'education', 'occupation_category', 'occupation', 'family_type', 'family_about', 'timeline', 'relocation', 'children', 'hobbies', 'values'];

export function completeness(d: Details, photos: number, horoscope: boolean) {
  const filled = IMPORTANT.filter(k => { const v = d[k]; return typeof v === 'number' ? v > 0 : !!(v && String(v).trim()) && !(k === 'profile_kind' && v === 'person'); }).length;
  const missing = IMPORTANT.filter(k => !d[k]);
  return { pct: Math.round(((filled + (photos > 0 ? 3 : 0) + (horoscope ? 2 : 0)) / (IMPORTANT.length + 5)) * 100), missing };
}

function Select({ field, value, onChange, label }: { field: string; value?: string; onChange: (v: string) => void; label: string }) {
  return <label className="field"><span>{label}</span><select value={value ?? ''} onChange={e => onChange(e.target.value)}><option value="">Not shared</option>{OPTIONS[field].map(([v, l]) => <option key={v} value={v}>{l}</option>)}</select></label>;
}

export function ProfileEditor({ onSaved }: { onSaved: () => void }) {
  const t = useT();
  const [d, setD] = useState<Details>(emptyBiodata), [photos, setPhotos] = useState<MatrimonyPhoto[]>([]);
  const [active, setActive] = useState(false), [consent, setConsent] = useState(false), [horoscope, setHoroscope] = useState(false);
  const [me, setMe] = useState<MyProfile>(), [birthPlace, setBirthPlace] = useState<Place>(), [msg, setMsg] = useState(''), [error, setError] = useState(''), [busy, setBusy] = useState(false);
  const [push, setPush] = useState(false), [selfie, setSelfie] = useState<File | null>(null), [selfieConsent, setSelfieConsent] = useState(false);
  async function load() {
    const [rows, ph, account] = await Promise.all([memberAPI<MyProfile[]>('/api/matrimony/me'), memberAPI<MatrimonyPhoto[]>('/api/matrimony/photos'), memberAPI<{ birth_place: { name: string; lat: number; lon: number; tz: string } | null }>('/api/me')]);
    const p = rows[0];
    if (p) { setMe(p); setActive(p.active); setHoroscope(p.horoscope_visible); setD({ ...emptyBiodata, ...p.details, social_links: p.details.social_links || [] }); }
    setPhotos(ph);
    if (account.birth_place) setBirthPlace({ name: account.birth_place.name, region: '', lat: account.birth_place.lat, lon: account.birth_place.lon, tz: account.birth_place.tz });
    setPush(await pushEnabled().catch(() => false));
  }
  useEffect(() => { void load().catch(e => setError(e.message)); }, []);
  const set = (p: Partial<Details>) => setD(x => ({ ...x, ...p }));
  async function act(fn: () => Promise<void>) { setBusy(true); setError(''); setMsg(''); try { await fn(); } catch (e) { setError(e instanceof Error ? e.message : 'Request failed'); } finally { setBusy(false); } }
  const { pct, missing } = completeness(d, photos.filter(p => p.published).length, horoscope);
  const text = (k: keyof Details, l: string, max: number, rows = 1) => <label className="field"><span>{l}</span>{rows > 1 ? <textarea maxLength={max} rows={rows} value={(d[k] as string) ?? ''} onChange={e => set({ [k]: e.target.value })} /> : <input maxLength={max} value={(d[k] as string) ?? ''} onChange={e => set({ [k]: e.target.value })} />}</label>;
  return (
    <form className="mat-editor" onSubmit={e => { e.preventDefault(); void act(async () => {
      await memberAPI('/api/matrimony/me', 'PUT', { active, consent, horoscope, details: cleanBiodata(d), photo_ids: photos.filter(p => p.published).map(p => p.id) });
      setConsent(false); setMsg(active ? (me?.hidden ? 'Saved; the profile is still hidden by moderation.' : 'Profile saved and discoverable.') : 'Saved privately. Discovery is paused.'); onSaved(); await load();
    }); }}>
      <section className="card mat-progress" aria-label={t('Profile completeness')}>
        <div className="mat-progress-row"><b>{t('Profile completeness')}: {pct}%</b><div className="meter" role="progressbar" aria-valuenow={pct} aria-valuemin={0} aria-valuemax={100}><i style={{ width: pct + '%' }} /></div></div>
        {missing.length > 0 && <p className="muted small">Complete profiles get more replies. Still empty: {missing.slice(0, 6).map(k => k.replace(/_/g, ' ')).join(', ')}{missing.length > 6 ? '…' : ''}.</p>}
        <p className="small"><a href="#biodata">{t('Print biodata')} / save as PDF →</a></p>
      </section>

      <section className="card"><h3>About</h3><div className="grid-2">
        {text('display_name', 'Display name', 100)}
        <label className="field"><span>Profile for</span><select value={d.profile_kind} onChange={e => set({ profile_kind: e.target.value })}><option value="person">A person seeking a partner</option><option value="bride">Bride</option><option value="groom">Groom</option></select></label>
        <div className="field-wide">{text('introduction', 'Introduction', 1000, 3)}</div>
        <label className="field"><span>{t('Height')} (cm)</span><input type="number" min={120} max={230} value={d.height_cm || ''} onChange={e => set({ height_cm: Number(e.target.value) || 0 })} /></label>
        <Select field="marital_status" label={t('Marital status')} value={d.marital_status} onChange={v => set({ marital_status: v })} />
        <Select field="religion" label={t('Religion')} value={d.religion} onChange={v => set({ religion: v })} />
        {text('community', 'Community / caste (optional)', 100)}
        <Select field="mother_tongue" label={t('Mother tongue')} value={d.mother_tongue} onChange={v => set({ mother_tongue: v })} />
        {text('languages', 'Other languages', 200)}
        <Select field="diet" label={t('Diet')} value={d.diet} onChange={v => set({ diet: v })} />
        <div className="field-wide city-search"><PlaceSearch label={t('City') + ' (where you live)'} value={d.city ? { name: d.city, region: d.region || '', lat: 0, lon: 0, tz: '' } : undefined} onChange={p => set({ city: p.name, region: p.region })} /></div>
      </div></section>

      <section className="card"><h3>{t('Education')} &amp; {t('Occupation')}</h3><div className="grid-2">
        <Select field="education_level" label={t('Education')} value={d.education_level} onChange={v => set({ education_level: v })} />
        {text('education', 'Field of study / college', 200)}
        <Select field="occupation_category" label={t('Occupation')} value={d.occupation_category} onChange={v => set({ occupation_category: v })} />
        {text('occupation', 'Job title / employer', 200)}
        <Select field="income_band" label={t('Income') + ' (per year)'} value={d.income_band} onChange={v => set({ income_band: v })} />
      </div></section>

      <section className="card"><h3>{t('Family')} &amp; plans</h3><div className="grid-2">
        <Select field="family_type" label={t('Family type')} value={d.family_type} onChange={v => set({ family_type: v })} />
        <Select field="timeline" label={t('Marriage timeline')} value={d.timeline} onChange={v => set({ timeline: v })} />
        <Select field="relocation" label={t('Relocation')} value={d.relocation} onChange={v => set({ relocation: v })} />
        <Select field="children" label={t('Children')} value={d.children} onChange={v => set({ children: v })} />
        <div className="field-wide">{text('family_about', 'About the family', 1000, 3)}</div>
        {text('hobbies', 'Hobbies and interests', 500, 2)}
        {text('values', 'Values that matter to you', 500, 2)}
        <label className="field"><span>Partner age from</span><input type="number" min={18} max={100} value={d.min_age} onChange={e => set({ min_age: +e.target.value })} /></label>
        <label className="field"><span>to</span><input type="number" min={18} max={100} value={d.max_age} onChange={e => set({ max_age: +e.target.value })} /></label>
        <label className="field field-wide"><span>Social profile links (optional, one per line)</span><textarea rows={2} value={(d.social_links || []).join('\n')} onChange={e => set({ social_links: e.target.value.split('\n') })} placeholder="https://…" /></label>
      </div></section>

      <section className="card"><PhotoEditor photos={photos} onChange={setPhotos} /></section>

      <section className="card mat-horoscope"><h3>Horoscope matching</h3>
        <p>When you and another member both turn this on, each of you sees your guna score (out of 36), Moon sign, nakshatra, Mangal dosha and root-number planet on the other's card, and can open the full kundali comparison with Ask Astrisk. <b>Your birth date, time and place are never shown</b> to anyone.</p>
        <label className="consent"><input type="checkbox" checked={horoscope} onChange={e => setHoroscope(e.target.checked)} /> <span>{t('Show horoscope matching')}</span></label>
        <div className="mat-birthplace">
          <PlaceSearch label={t('Birthplace') + ' (needed for Mangal dosha and lagna)'} value={birthPlace} onChange={p => void act(async () => { await memberAPI('/api/me/birth-place', 'PUT', { name: [p.name, p.region].filter(Boolean).join(', '), lat: p.lat, lon: p.lon, tz: p.tz }); setBirthPlace(p); setMsg('Birthplace saved.'); })} />
          {!birthPlace && <p className="muted small">Without a birthplace, the guna score still works (it uses the Moon), but Mangal dosha shows as unknown.</p>}
        </div>
      </section>

      <section className="card mat-verify"><h3>Photo verification {me?.verified && <span className="verified">✓ {t('Verified')}</span>}</h3>
        {me?.verified ? <p>Your profile shows a verified badge. Uploading new photos keeps it; a moderator may re-check after reports.</p> : me?.verification === 'pending' ? <p>Your selfie is waiting for a moderator. It is deleted after review.</p> : <>
          <p>Take a selfie now and upload it. A moderator compares it with your published photos and then deletes it; only the badge remains.{me?.verification === 'rejected' && ' Your last selfie could not be matched; please try again with good light.'}</p>
          <div className="field-row">
            <label className="field"><span>Selfie (JPEG or PNG)</span><input type="file" accept="image/jpeg,image/png" capture="user" onChange={e => setSelfie(e.target.files?.[0] ?? null)} /></label>
            <label className="consent"><input type="checkbox" checked={selfieConsent} onChange={e => setSelfieConsent(e.target.checked)} /> <span>A moderator may view this selfie to verify me.</span></label>
            <button type="button" className="ghost" disabled={busy || !selfie || !selfieConsent} onClick={() => void act(async () => { const f = new FormData(); f.append('photo', selfie!); f.append('consent', 'true'); await memberAPI('/api/matrimony/verification', 'POST', f); setMsg('Selfie sent for verification.'); await load(); })}>Send for verification</button>
          </div></>}
      </section>

      <section className="card mat-alerts"><h3>Alerts</h3>
        <label className="consent"><input type="checkbox" checked={me?.email_alerts ?? true} onChange={e => void act(async () => { await memberAPI('/api/me/alerts', 'PUT', { email: e.target.checked }); await load(); })} /> <span>Email me about new interests, accepted interests and messages (needs a verified email).</span></label>
        {pushSupported() ? <p><button type="button" className="ghost" disabled={busy} onClick={() => void act(async () => { if (push) { await disablePush(); setPush(false); setMsg('Phone notifications turned off.'); } else { await enablePush(); setPush(true); setMsg('Phone notifications turned on for this device.'); } })}>{push ? 'Turn off notifications on this device' : 'Turn on notifications on this device'}</button></p>
          : <p className="muted small">This browser cannot show notifications. On iPhone, open Astrisk in Safari, choose Share → Add to Home Screen, then turn notifications on from there.</p>}
      </section>

      <section className="card mat-publish"><h3>Visibility</h3>
        {me?.hidden && <p role="status">Your profile is hidden by moderation. Editing does not restore visibility.</p>}
        <label className="consent"><input type="checkbox" checked={active} onChange={e => setActive(e.target.checked)} /> <span>Make my profile discoverable to eligible members</span></label>
        {active && <label className="consent"><input type="checkbox" required checked={consent} onChange={e => setConsent(e.target.checked)} /> <span>I am an adult, these details are accurate, and I consent to sharing them with participating members.</span></label>}
        <button className="primary" disabled={busy}>{t('Save profile')}</button>
        {msg && <p role="status" className="good-text">{msg}</p>}
        {error && <p role="alert" className="form-error">{error}</p>}
      </section>
    </form>
  );
}
