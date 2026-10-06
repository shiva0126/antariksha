import { useEffect, useState } from 'react';
import { ButtonLink, Card, EmptyState, PageHeader, Skeleton, Tabs } from '../../ds';
import { useT } from '../../i18n';
import { api } from '../../lib/api';
import { href, pick } from '../../lib/router';
import { Feed } from './Feed';
import { People } from './People';
import { FamilySpace } from './FamilySpace';
import './community.css';

const TABS = ['feed', 'people', 'family'] as const;

/** Community: feed, people and private family trees. Profile, notifications,
 *  security and moderation live in the Me menu and Admin. */
export function CommunityPage({ sub = '' }: { sub?: string }) {
  const t = useT();
  const tab = pick(sub, TABS);
  const [joined, setJoined] = useState<boolean>();
  useEffect(() => { api<{ community: boolean }[]>('/api/community/settings').then(r => setJoined(!!r[0]?.community)).catch(() => setJoined(false)); }, []);
  return (
    <div className="page community-page">
      <PageHeader kicker="People · stories · belonging" title={t('Community')} compactOnMobile description="Share your world, find common ground and keep a private family tree. You choose who sees what." />
      <Tabs label="Community" base="community" active={tab} items={[{ id: 'feed', label: t('Feed') }, { id: 'people', label: t('People') }, { id: 'family', label: t('Family tree') }]} />
      {joined === undefined ? <Card><Skeleton /></Card> : !joined && tab !== 'family' ? (
        <Card><EmptyState title="Join the community to share and follow" action={<ButtonLink variant="primary" href={href('me', 'profile')}>Set up my community profile</ButtonLink>}>
          Your community profile is off. Turn it on in your profile with your adult date of birth; until then your posts can still be saved privately.
        </EmptyState></Card>
      ) : null}
      {tab === 'feed' && <Feed />}
      {tab === 'people' && <People />}
      {tab === 'family' && <FamilySpace />}
    </div>
  );
}
