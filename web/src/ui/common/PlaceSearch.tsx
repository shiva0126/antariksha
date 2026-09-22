import { useEffect, useId, useRef, useState } from 'react';
import { searchPlaces } from '../../api/client';
import { places as builtIn, placeLabel, type Place } from './places';

/** Birthplace combobox: searches ~34,000 cities on the server, falling back to
 * the built-in list if the server can't be reached. */
export function PlaceSearch({ value, onChange, label = 'Birthplace', compact }: { value?: Place; onChange: (p: Place) => void; label?: string; compact?: boolean }) {
  const [text, setText] = useState(value ? placeLabel(value) : '');
  const [hits, setHits] = useState<Place[]>([]);
  const [open, setOpen] = useState(false);
  const [cursor, setCursor] = useState(0);
  const id = useId();
  const box = useRef<HTMLDivElement>(null);

  useEffect(() => { if (value) setText(placeLabel(value)); }, [value]);

  useEffect(() => {
    const q = text.trim();
    if (!open || q.length < 2 || (value && q === placeLabel(value))) { setHits([]); return; }
    const ctrl = new AbortController();
    const timer = setTimeout(() => {
      searchPlaces(q, ctrl.signal)
        .then(r => setHits(r.places.map(p => ({ name: p.name, region: [p.region, p.country].filter(Boolean).join(', '), lat: p.lat, lon: p.lon, tz: p.tz }))))
        .catch(e => { if (e.name !== 'AbortError') setHits(builtIn.filter(p => placeLabel(p).toLowerCase().includes(q.toLowerCase())).slice(0, 8)); });
      setCursor(0);
    }, 180);
    return () => { clearTimeout(timer); ctrl.abort(); };
  }, [text, open, value]);

  useEffect(() => {
    const close = (e: MouseEvent) => { if (!box.current?.contains(e.target as Node)) setOpen(false); };
    document.addEventListener('mousedown', close);
    return () => document.removeEventListener('mousedown', close);
  }, []);

  function pick(p: Place) { onChange(p); setText(placeLabel(p)); setOpen(false); setHits([]); }

  return (
    <div className={'field place-search' + (compact ? ' compact' : '')} ref={box}>
      <label htmlFor={id}><span>{label}</span></label>
      <input id={id} role="combobox" aria-expanded={open && hits.length > 0} aria-controls={id + '-list'} aria-autocomplete="list" autoComplete="off"
        value={text} placeholder="Search any city or town" onFocus={() => setOpen(true)}
        onChange={e => { setText(e.target.value); setOpen(true); }}
        onKeyDown={e => {
          if (!hits.length) return;
          if (e.key === 'ArrowDown') { e.preventDefault(); setCursor(c => Math.min(c + 1, hits.length - 1)); }
          else if (e.key === 'ArrowUp') { e.preventDefault(); setCursor(c => Math.max(c - 1, 0)); }
          else if (e.key === 'Enter') { e.preventDefault(); pick(hits[cursor]); }
          else if (e.key === 'Escape') setOpen(false);
        }} />
      {open && hits.length > 0 && (
        <ul id={id + '-list'} role="listbox" className="place-list">
          {hits.map((p, i) => (
            <li key={`${p.name}${p.lat}${p.lon}`} role="option" aria-selected={i === cursor} className={i === cursor ? 'active' : ''}
              onMouseDown={e => { e.preventDefault(); pick(p); }} onMouseEnter={() => setCursor(i)}>
              <b>{p.name}</b><span>{p.region}</span><em>{p.tz}</em>
            </li>
          ))}
        </ul>
      )}
      {!compact && <small>{value ? `${value.lat.toFixed(4)}°, ${value.lon.toFixed(4)}° · ${value.tz}` : 'Type at least two letters'}</small>}
    </div>
  );
}
