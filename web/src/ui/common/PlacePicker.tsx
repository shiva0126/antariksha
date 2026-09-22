import { useState } from 'react';
import { findPlace, placeLabel, places, type Place } from './places';

export function PlacePicker({ value, onChange }: { value: Place; onChange: (p: Place) => void }) {
  const [text, setText] = useState(placeLabel(value));
  return (
    <label className="field compact">
      <span>Location</span>
      <input list="panchang-places" value={text} aria-label="Location"
        onChange={e => { setText(e.target.value); const p = findPlace(e.target.value); if (p) onChange(p); }}
        onBlur={() => { if (!findPlace(text)) setText(placeLabel(value)); }} />
      <datalist id="panchang-places">{places.map(p => <option key={placeLabel(p)} value={placeLabel(p)} />)}</datalist>
    </label>
  );
}
