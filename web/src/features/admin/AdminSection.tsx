import { PageHeader, Tabs } from '../../ds';
import { pick } from '../../lib/router';
import { AdminPage } from '../../ui/AdminPage';
import { Moderation } from '../../ui/community/Settings';
import { MatrimonyModeration, VerificationQueue } from '../../ui/community/MatrimonyModeration';

/** One admin area: the superadmin console, or reports and photo
 *  verification for moderators. */
export function AdminSection({ sub, role }: { sub: string; role: string }) {
  if (role === 'superadmin') return <AdminPage />;
  if (role !== 'moderator') return <div className="page"><PageHeader title="Admin access required" description="This area is for moderators and the superadmin." /></div>;
  const tabs = ['reports', 'verifications'] as const;
  const tab = pick(sub, tabs);
  const labels = { reports: 'Reports', verifications: 'Photo verification' };
  return (
    <div className="page">
      <PageHeader kicker="Moderation" title={labels[tab]} />
      <Tabs label="Moderation" base="admin" active={tab} items={tabs.map(id => ({ id, label: labels[id] }))} />
      {tab === 'reports' && <><Moderation /><MatrimonyModeration /></>}
      {tab === 'verifications' && <VerificationQueue />}
    </div>
  );
}
