import { useEffect, useState } from 'react';
import { getChart } from '../../api/client';
import type { ChartResponse } from '../../api/types';
import { NorthIndian } from '../../chart2d/NorthIndian';
import { SouthIndian } from '../../chart2d/SouthIndian';
import { useNames, useT } from '../../i18n';
import { memberAPI } from '../community/shared';
import type { MatrimonyPhoto } from '../community/Biodata';
import { PlanetTable } from '../kundali/PlanetTable';
import { heightLabel, label } from './options';
import type { MyProfile } from './types';
import './matrimony.css';

type Me = { birth_date: string; birth_time: string; birth_place: { name: string; lat: number; lon: number; tz: string } | null; profile: { name?: string } };

/** A printable biodata in the format families share, with an optional
 *  kundali page. "Print / save as PDF" uses the browser's print dialog. */
export function BiodataPrint() {
  const t = useT(), n = useNames();
  const [p, setP] = useState<MyProfile>(), [photos, setPhotos] = useState<MatrimonyPhoto[]>([]), [me, setMe] = useState<Me>(), [chart, setChart] = useState<ChartResponse>();
  const [withKundali, setWithKundali] = useState(true), [withBirth, setWithBirth] = useState(true), [error, setError] = useState('');
  useEffect(() => {
    Promise.all([memberAPI<MyProfile[]>('/api/matrimony/me'), memberAPI<MatrimonyPhoto[]>('/api/matrimony/photos'), memberAPI<Me>('/api/me')]).then(([rows, ph, m]) => {
      if (!rows[0]) { setError('Create your matrimony profile first.'); return; }
      setP(rows[0]); setPhotos(ph.filter(x => x.published)); setMe(m);
      if (m.birth_place && m.birth_date && m.birth_time) getChart({ date: m.birth_date, time: m.birth_time, lat: m.birth_place.lat, lon: m.birth_place.lon, tz: m.birth_place.tz }).then(setChart).catch(() => {});
    }).catch(e => setError(e.message));
  }, []);
  if (error) return <div className="page"><p role="alert" className="form-error">{error}</p><a href="#matrimony/profile">← Back to my profile</a></div>;
  if (!p || !me) return <div className="page"><p className="muted">Preparing your biodata…</p></div>;
  const d = p.details;
  const moon = chart?.grahas.find(g => g.id === 'moon');
  const rows: [string, string | undefined][] = [
    ['Name', d.display_name || me.profile?.name], ['Age', me.birth_date ? String(Math.floor((Date.now() - Date.parse(me.birth_date)) / 31557600000)) : ''],
    ...(withBirth ? [['Birth date', me.birth_date], ['Birth time', me.birth_time], ['Birthplace', me.birth_place?.name]] as [string, string][] : []),
    ['Height', heightLabel(d.height_cm)], ['Marital status', label('marital_status', d.marital_status)], ['Religion', [label('religion', d.religion), d.community].filter(Boolean).join(' · ')],
    ['Mother tongue', label('mother_tongue', d.mother_tongue)], ['Languages', d.languages], ['Diet', label('diet', d.diet)], ['City', [d.city, d.region].filter(Boolean).join(', ')],
    ['Education', [label('education_level', d.education_level), d.education].filter(Boolean).join(' · ')], ['Occupation', [d.occupation, label('occupation_category', d.occupation_category)].filter(Boolean).join(' · ')],
    ['Income', label('income_band', d.income_band)], ['Family type', label('family_type', d.family_type)], ['Marriage timeline', label('timeline', d.timeline)],
    ['Relocation', label('relocation', d.relocation)], ['Children', label('children', d.children)], ['Hobbies', d.hobbies], ['Values', d.values],
    ...(moon ? [['Rashi (Moon sign)', n(moon.rashi)], ['Nakshatra', `${n(moon.nakshatra)}, pada ${moon.nakshatra_pada}`], ['Lagna', n(chart!.ascendant.rashi)]] as [string, string][] : []),
  ];
  return (
    <article className="biodata-print">
      <div className="no-print biodata-controls card">
        <a href="#matrimony/profile">← Back to my profile</a>
        <label className="consent"><input type="checkbox" checked={withBirth} onChange={e => setWithBirth(e.target.checked)} /> <span>Include birth date, time and place</span></label>
        <label className="consent"><input type="checkbox" checked={withKundali} disabled={!chart} onChange={e => setWithKundali(e.target.checked)} /> <span>Include kundali page{!chart && ' (add your birthplace in My matrimony profile)'}</span></label>
        <button className="primary" onClick={() => window.print()}>{t('Print biodata')} / save as PDF</button>
        <p className="muted small">The PDF is created on your device. Share it only with families you trust.</p>
      </div>
      <section className="biodata-sheet">
        <header><p className="biodata-om">॥ Biodata ॥</p><h1>{d.display_name || me.profile?.name || 'Biodata'}</h1></header>
        <div className="biodata-top">
          {photos[0] && <img src={`/api/matrimony/photos/${photos[0].id}`} alt={photos[0].alt || 'Profile photo'} />}
          <table><tbody>{rows.filter(([, v]) => v).map(([k, v]) => <tr key={k}><th>{t(k)}</th><td>{v}</td></tr>)}</tbody></table>
        </div>
        {d.introduction && <><h2>About me</h2><p>{d.introduction}</p></>}
        {d.family_about && <><h2>About the family</h2><p>{d.family_about}</p></>}
        {p.verified && <p className="small">✓ Photo verified on Astrisk</p>}
      </section>
      {withKundali && chart && (
        <section className="biodata-sheet biodata-kundali">
          <h2>Janma kundali</h2>
          <p className="small">Lahiri ayanamsa · whole-sign houses · computed by Astrisk (Swiss Ephemeris)</p>
          <div className="report-charts"><figure><SouthIndian chart={chart} /><figcaption>South Indian</figcaption></figure><figure><NorthIndian chart={chart} /><figcaption>North Indian</figcaption></figure></div>
          <PlanetTable chart={chart} />
        </section>
      )}
    </article>
  );
}
