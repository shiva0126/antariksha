import { PageHeader, Tabs } from '../../ds';
import { useT } from '../../i18n';
import { pick } from '../../lib/router';
import { ProfileTab } from './ProfileTab';
import { ChartsTab, NotificationsTab, PrivacyTab, SecurityTab, SettingsTab } from './OtherTabs';
import './me.css';

const TABS = ['profile', 'charts', 'notifications', 'security', 'privacy', 'settings'] as const;

/** Everything about the member in one place (replaces My profile, Profile &
 *  interests, My character, Security and the Privacy page). */
export function MePage({ sub }: { sub: string }) {
  const t = useT();
  const tab = pick(sub, TABS);
  const labels: Record<typeof TABS[number], string> = { profile: t('Profile'), charts: t('Saved charts'), notifications: t('Notifications'), security: t('Security'), privacy: t('Privacy and data'), settings: t('Settings') };
  return (
    <div className="page">
      <PageHeader kicker="Your private space" title={labels[tab]} />
      <Tabs label="Me" base="me" active={tab} items={TABS.map(id => ({ id, label: labels[id] }))} />
      {tab === 'profile' && <ProfileTab />}
      {tab === 'charts' && <ChartsTab />}
      {tab === 'notifications' && <NotificationsTab />}
      {tab === 'security' && <SecurityTab />}
      {tab === 'privacy' && <PrivacyTab />}
      {tab === 'settings' && <SettingsTab />}
    </div>
  );
}
