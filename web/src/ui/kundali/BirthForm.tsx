import { useMemo, useState } from 'react';
import type { ChartInput } from '../../api/types';
import { defaultPlace, findPlace, nearestPlace, placeLabel, places, timeZones } from '../common/places';

export interface BirthDetails extends ChartInput { name: string; place: string }

export function BirthForm({ onSubmit, busy, initial }: { onSubmit: (b: BirthDetails) => void; busy: boolean; initial?: BirthDetails }) {
  const initialPlace = initial ? nearestPlace(initial.lat, initial.lon) : defaultPlace;
  const [custom, setCustom] = useState(Boolean(initial && !initialPlace));
  const [placeText, setPlaceText] = useState(initialPlace ? placeLabel(initialPlace) : '');
  const [error, setError] = useState('');
  const zones = useMemo(timeZones, []);
  const match = findPlace(placeText);

  function submit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const f = new FormData(e.currentTarget);
    const base = { name: String(f.get('name') || '').trim(), date: String(f.get('date')), time: String(f.get('time')) };
    if (custom) {
      const lat = Number(f.get('lat')), lon = Number(f.get('lon')), tz = String(f.get('tz'));
      if (!Number.isFinite(lat) || lat < -90 || lat > 90 || !Number.isFinite(lon) || lon < -180 || lon > 180) return setError('Latitude must be −90 to 90 and longitude −180 to 180.');
      if (!zones.includes(tz)) return setError('Choose a timezone from the list.');
      setError('');
      return onSubmit({ ...base, lat, lon, tz, place: `${lat.toFixed(4)}°, ${lon.toFixed(4)}°` });
    }
    if (!match) return setError('Pick a birthplace from the list, or enter coordinates.');
    setError('');
    onSubmit({ ...base, lat: match.lat, lon: match.lon, tz: match.tz, place: placeLabel(match) });
  }

  return (
    <form className="birth-form" onSubmit={submit} noValidate>
      <label className="field field-wide">
        <span>Name <em>(optional)</em></span>
        <input name="name" autoComplete="name" placeholder="Whose chart is this?" defaultValue={initial?.name} maxLength={60} />
      </label>
      <label className="field">
        <span>Birth date</span>
        <input required name="date" type="date" min="1900-01-01" max="2100-12-31" defaultValue={initial?.date ?? '1996-05-14'} />
      </label>
      <label className="field">
        <span>Birth time</span>
        <input required name="time" type="time" defaultValue={initial?.time ?? '10:15'} />
        <small>Local clock time as recorded, including daylight saving.</small>
      </label>
      {!custom ? (
        <label className="field field-wide">
          <span>Birthplace</span>
          <input list="birth-places" value={placeText} onChange={e => setPlaceText(e.target.value)} placeholder="Start typing a city" autoComplete="off" aria-invalid={!match && placeText !== ''} />
          <datalist id="birth-places">{places.map(p => <option key={placeLabel(p)} value={placeLabel(p)} />)}</datalist>
          <small>{match ? `${match.lat.toFixed(4)}°, ${match.lon.toFixed(4)}° · ${match.tz}` : 'Not in the list? Use coordinates instead.'}</small>
        </label>
      ) : (
        <div className="field-row field-wide">
          <label className="field"><span>Latitude</span><input name="lat" type="number" step="any" min="-90" max="90" required defaultValue={initial?.lat ?? 12.9716} /></label>
          <label className="field"><span>Longitude</span><input name="lon" type="number" step="any" min="-180" max="180" required defaultValue={initial?.lon ?? 77.5946} /></label>
          <label className="field"><span>Timezone</span>
            <select name="tz" defaultValue={initial?.tz ?? 'Asia/Kolkata'}>{zones.map(z => <option key={z}>{z}</option>)}</select>
          </label>
        </div>
      )}
      <button type="button" className="link-button field-wide" onClick={() => { setCustom(!custom); setError(''); }}>
        {custom ? '← Choose a city instead' : 'Enter latitude, longitude and timezone instead'}
      </button>
      {error && <p role="alert" className="form-error field-wide">{error}</p>}
      <button className="primary field-wide" disabled={busy}>{busy ? 'Calculating your chart…' : 'Reveal my chart'}</button>
    </form>
  );
}
