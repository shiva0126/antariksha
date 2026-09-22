import { useEffect, useState } from 'react';
import { getVarga } from '../../api/client';
import type { ChartResponse, ReadingResponse } from '../../api/types';
import { NorthIndian } from '../../chart2d/NorthIndian';
import { SouthIndian } from '../../chart2d/SouthIndian';
import { longDate } from '../../astro/format';
import type { Profile } from '../common/profiles';
import { DashaTimeline } from './DashaTimeline';
import { PlanetTable } from './PlanetTable';
import { ReadingPanel } from './ReadingPanel';

/** A complete, printable kundali report. "Print / save PDF" opens the
 * browser's print dialog, which can save it as a PDF. Free, no sign-up. */
export function ReportView({ profile, chart, reading }: { profile: Profile; chart: ChartResponse; reading?: ReadingResponse }) {
  const [d9, setD9] = useState<ChartResponse>();
  useEffect(() => { getVarga(profile, 9).then(r => setD9(r.chart)).catch(() => {}); }, [profile]);
  return (
    <article className="report">
      <header className="report-head">
        <p className="kicker">Antariksha · Janma kundali report</p>
        <h1>{profile.name || 'Birth chart'}</h1>
        <p>{longDate(profile.date)} at {profile.time} · {profile.place} ({profile.lat.toFixed(4)}°, {profile.lon.toFixed(4)}°, {profile.tz})</p>
        <p className="muted small">Lahiri ayanamsa · whole-sign houses · Swiss Ephemeris · generated {new Date().toLocaleDateString()}</p>
        <button className="primary no-print" onClick={() => window.print()}>Print / save as PDF</button>
      </header>
      <section className="report-charts">
        <figure><SouthIndian chart={chart} /><figcaption>Rashi (D1), South Indian</figcaption></figure>
        <figure><NorthIndian chart={chart} /><figcaption>Rashi (D1), North Indian</figcaption></figure>
        {d9 && <figure><SouthIndian chart={d9} title="Navamsha · D9" /><figcaption>Navamsha (D9)</figcaption></figure>}
      </section>
      <section><h2>Planetary positions</h2><PlanetTable chart={chart} facts={reading?.facts} /></section>
      {reading && <section><h2>Dashas</h2><DashaTimeline facts={reading.facts} /></section>}
      {reading && <section><h2>Reading</h2><ReadingPanel data={reading} /></section>}
    </article>
  );
}
