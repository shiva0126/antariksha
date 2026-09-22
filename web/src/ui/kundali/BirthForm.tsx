import { useMemo, useState } from 'react';
import type { ChartInput } from '../../api/types';
import { useT } from '../../i18n';
import { PlaceSearch } from '../common/PlaceSearch';
import { defaultPlace, nearestPlace, placeLabel, timeZones, type Place } from '../common/places';

export interface BirthDetails extends ChartInput { name: string; place: string }

export interface BirthDraft { name: string; date: string; time: string; place?: Place; custom: boolean; lat: string; lon: string; tz: string }

export const draftFrom = (b?: BirthDetails): BirthDraft => {
  const near = b ? nearestPlace(b.lat, b.lon) : defaultPlace;
  const place = near ?? (b ? { name: b.place, region: '', lat: b.lat, lon: b.lon, tz: b.tz } : undefined);
  return { name: b?.name ?? '', date: b?.date ?? '1996-05-14', time: b?.time ?? '10:15', place, custom: false, lat: String(b?.lat ?? ''), lon: String(b?.lon ?? ''), tz: b?.tz ?? 'Asia/Kolkata' };
};

/** Validates a draft; returns details or an error message. */
export function resolveDraft(d: BirthDraft): BirthDetails | string {
  if (!/^\d{4}-\d{2}-\d{2}$/.test(d.date)) return 'Enter the birth date.';
  if (!/^\d{2}:\d{2}$/.test(d.time)) return 'Enter the birth time.';
  const base = { name: d.name.trim(), date: d.date, time: d.time };
  if (d.custom) {
    const lat = Number(d.lat), lon = Number(d.lon);
    if (!d.lat || !Number.isFinite(lat) || lat < -90 || lat > 90 || !d.lon || !Number.isFinite(lon) || lon < -180 || lon > 180) return 'Latitude must be −90 to 90 and longitude −180 to 180.';
    return { ...base, lat, lon, tz: d.tz, place: `${lat.toFixed(4)}°, ${lon.toFixed(4)}°` };
  }
  if (!d.place) return 'Pick a birthplace from the suggestions, or enter coordinates.';
  return { ...base, lat: d.place.lat, lon: d.place.lon, tz: d.place.tz, place: placeLabel(d.place) };
}

export function BirthFields({ draft, onChange, legend, showName = true }: { draft: BirthDraft; onChange: (d: BirthDraft) => void; legend?: string; showName?: boolean }) {
  const t = useT();
  const zones = useMemo(timeZones, []);
  const set = (p: Partial<BirthDraft>) => onChange({ ...draft, ...p });
  return (
    <fieldset className="birth-fields">
      {legend && <legend>{legend}</legend>}
      {showName && <label className="field field-wide"><span>{t('Name')} <em>(optional)</em></span>
        <input value={draft.name} onChange={e => set({ name: e.target.value })} autoComplete="name" placeholder="Whose chart is this?" maxLength={60} /></label>}
      <label className="field"><span>{t('Birth date')}</span>
        <input type="date" required min="1900-01-01" max="2100-12-31" value={draft.date} onChange={e => set({ date: e.target.value })} /></label>
      <label className="field"><span>{t('Birth time')}</span>
        <input type="time" required value={draft.time} onChange={e => set({ time: e.target.value })} />
        <small>Local clock time as recorded, including daylight saving.</small></label>
      {!draft.custom ? (
        <div className="field-wide"><PlaceSearch label={t('Birthplace')} value={draft.place} onChange={p => set({ place: p })} /></div>
      ) : (
        <div className="field-row field-wide">
          <label className="field"><span>Latitude</span><input type="number" step="any" value={draft.lat} onChange={e => set({ lat: e.target.value })} placeholder="12.9716" /></label>
          <label className="field"><span>Longitude</span><input type="number" step="any" value={draft.lon} onChange={e => set({ lon: e.target.value })} placeholder="77.5946" /></label>
          <label className="field"><span>Timezone</span><select value={draft.tz} onChange={e => set({ tz: e.target.value })}>{zones.map(z => <option key={z}>{z}</option>)}</select></label>
        </div>
      )}
      <button type="button" className="link-button field-wide" onClick={() => set({ custom: !draft.custom })}>
        {draft.custom ? '← Search for a city instead' : 'Enter latitude, longitude and timezone instead'}
      </button>
    </fieldset>
  );
}

export function BirthForm({ onSubmit, busy, initial, submitLabel }: { onSubmit: (b: BirthDetails) => void; busy: boolean; initial?: BirthDetails; submitLabel?: string }) {
  const t = useT();
  const [draft, setDraft] = useState(() => draftFrom(initial));
  const [consent, setConsent] = useState(Boolean(initial));
  const [error, setError] = useState('');
  function submit(e: React.FormEvent) {
    e.preventDefault();
    const r = resolveDraft(draft);
    if (typeof r === 'string') return setError(r);
    if (!consent) return setError('Please confirm the consent statement to continue.');
    setError('');
    onSubmit(r);
  }
  return (
    <form className="birth-form" onSubmit={submit} noValidate>
      <BirthFields draft={draft} onChange={setDraft} />
      <label className="consent field-wide">
        <input type="checkbox" checked={consent} onChange={e => setConsent(e.target.checked)} />
        <span>I agree that Antariksha may use these birth details to calculate the chart and may store questions I ask so I can see my history. I am 18 or older, or have a parent's permission. Everything is free, and I can delete my data at any time. <a href="#privacy">Privacy</a></span>
      </label>
      {error && <p role="alert" className="form-error field-wide">{error}</p>}
      <button className="primary field-wide" disabled={busy}>{busy ? 'Calculating your chart…' : submitLabel ?? t('Reveal my chart')}</button>
    </form>
  );
}
