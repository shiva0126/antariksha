import { useState } from 'react';
import type { ChartResponse, ReadingResponse } from '../../api/types';
import { grahaEnglish, longDate } from '../../astro/format';
import { useNames, useT } from '../../i18n';
import type { Profile } from '../common/profiles';

interface Props {
  profile: Profile; profiles: Profile[]; chart: ChartResponse; reading?: ReadingResponse;
  onEdit: () => void; onSwitch: (id: string) => void; onAdd: () => void; onDelete: (id: string) => void; onPrint: () => void;
}

export function ProfileBar({ profile, profiles, chart, reading, onEdit, onSwitch, onAdd, onDelete, onPrint }: Props) {
  const t = useT(), n = useNames();
  const [confirm, setConfirm] = useState(false);
  const moon = chart.grahas.find(g => g.id === 'moon');
  const sun = chart.grahas.find(g => g.id === 'sun');
  const dasha = reading?.facts.vimshottari.current;
  const facts: [string, string][] = [
    [t('Lagna'), `${n(chart.ascendant.rashi)} ${chart.ascendant.degree.toFixed(1)}°`],
    [t('Moon sign'), moon ? n(moon.rashi) : '—'],
    [t('Nakshatra'), moon ? `${n(moon.nakshatra)} · pada ${moon.nakshatra_pada}` : '—'],
    [t('Sun sign'), sun ? n(sun.rashi) : '—'],
    [t('Current dasha'), dasha?.maha ? `${grahaEnglish(dasha.maha)} – ${grahaEnglish(dasha.antara ?? '')}` : reading ? '—' : '…'],
  ];
  return (
    <section className="profile card" aria-label="Birth details">
      <div className="profile-id">
        <div className="profile-switch">
          <label className="sr-only" htmlFor="profile-select">{t('Profiles')}</label>
          <select id="profile-select" value={profile.id} onChange={e => e.target.value === '__new' ? onAdd() : onSwitch(e.target.value)}>
            {profiles.map(p => <option key={p.id} value={p.id}>{p.name || `Chart of ${longDate(p.date)}`}</option>)}
            <option value="__new">+ {t('Add profile')}</option>
          </select>
        </div>
        <h1>{profile.name || t('Your chart')}</h1>
        <p className="muted">{longDate(profile.date)} · {profile.time} · {profile.place} · {profile.tz}</p>
        <div className="profile-actions">
          <button className="ghost" onClick={onEdit}>{t('Edit birth details')}</button>
          <button className="ghost" onClick={onPrint}>Print / save PDF</button>
          {!confirm ? <button className="ghost danger" onClick={() => setConfirm(true)}>Delete</button> : (
            <span className="confirm">Delete this profile and its saved questions? <button className="ghost danger" onClick={() => { setConfirm(false); onDelete(profile.id); }}>Yes, delete</button> <button className="ghost" onClick={() => setConfirm(false)}>Cancel</button></span>
          )}
        </div>
      </div>
      <dl className="profile-facts">
        {facts.map(([k, v]) => <div key={k}><dt>{k}</dt><dd>{v}</dd></div>)}
      </dl>
    </section>
  );
}
