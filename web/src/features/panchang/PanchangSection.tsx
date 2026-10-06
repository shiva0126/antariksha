import { Tabs } from '../../ds';
import { useT } from '../../i18n';
import { pick } from '../../lib/router';
import { MuhurtaPage } from '../../ui/muhurta/MuhurtaPage';
import { PanchangPage } from '../../ui/panchang/PanchangPage';

const TABS = ['today', 'calendar', 'muhurta'] as const;

/** Panchang: today's five limbs, the month calendar and muhurta, as tabs. */
export function PanchangSection({ sub }: { sub: string }) {
  const t = useT();
  const tab = pick(sub, TABS);
  const tabs = <Tabs label="Panchang" base="panchang" active={tab} items={[{ id: 'today', label: t('Today') }, { id: 'calendar', label: t('Calendar') }, { id: 'muhurta', label: t('Muhurta') }]} />;
  if (tab === 'muhurta') return <MuhurtaPage tabs={tabs} />;
  return <PanchangPage key={tab} mode={tab === 'calendar' ? 'month' : 'day'} tabs={tabs} />;
}
