import type { ChartResponse, ReadingResponse } from '../../api/types';
import { grahaEnglish, longDate } from '../../astro/format';
import { Button, Card, Chip, StatTile, useFeedback } from '../../ds';
import { useNames, useT } from '../../i18n';
import { href } from '../../lib/router';
import type { Profile } from '../common/profiles';

/** The compact bar above every Kundali tab: which chart, its birth line, key
 *  facts, and chart actions. Account sync lives in Me → Saved charts. */
export function ChartBar({ profile, profiles, chart, reading, onEdit, onSwitch, onAdd, onDelete, onPrint }: {
  profile: Profile; profiles: Profile[]; chart: ChartResponse; reading?: ReadingResponse;
  onEdit: () => void; onSwitch: (id: string) => void; onAdd: () => void; onDelete: (id: string) => void; onPrint: () => void;
}) {
  const t = useT(), n = useNames();
  const { confirm } = useFeedback();
  const moon = chart.grahas.find(g => g.id === 'moon'), sun = chart.grahas.find(g => g.id === 'sun');
  const dasha = reading?.facts.vimshottari.current;
  return (
    <Card className="chart-bar" aria-label="Birth details">
      <div className="chart-bar__who">
        <label className="chart-bar__switch">
          <span className="ds-field__label">{t('Profiles')}</span>
          <select id="profile-select" value={profile.id} onChange={e => e.target.value === '__new' ? onAdd() : onSwitch(e.target.value)}>
            {profiles.map(p => <option key={p.id} value={p.id}>{p.name || `Chart of ${longDate(p.date)}`}</option>)}
            <option value="__new">+ {t('Add profile')}</option>
          </select>
        </label>
        <h1>{profile.name || t('Your chart')}</h1>
        <p className="ds-muted">{longDate(profile.date)} · {profile.time} · {profile.place}</p>
        <div className="ds-row">
          <Chip tone="neutral"><a href={href('me', 'charts')}>Saved charts</a></Chip>
          <Button size="sm" variant="ghost" onClick={onEdit}>{t('Edit birth details')}</Button>
          <Button size="sm" variant="ghost" onClick={onPrint}>Print / save PDF</Button>
          <Button size="sm" variant="ghost" className="danger-text" onClick={async () => { if (await confirm({ title: 'Delete this chart?', body: 'The chart and its saved questions are removed from this device.', confirm: 'Delete', danger: true })) onDelete(profile.id); }}>Delete</Button>
        </div>
      </div>
      <dl className="chart-bar__facts">
        <StatTile label={t('Lagna')} value={`${n(chart.ascendant.rashi)} ${chart.ascendant.degree.toFixed(1)}°`} />
        <StatTile label={t('Moon sign')} value={moon ? n(moon.rashi) : '—'} />
        <StatTile label={t('Nakshatra')} value={moon ? `${n(moon.nakshatra)} · ${moon.nakshatra_pada}` : '—'} />
        <StatTile label={t('Sun sign')} value={sun ? n(sun.rashi) : '—'} />
        <StatTile label={t('Current dasha')} value={dasha?.maha ? `${grahaEnglish(dasha.maha)} · ${grahaEnglish(dasha.antara ?? '')}` : reading ? '—' : '…'} />
      </dl>
    </Card>
  );
}
