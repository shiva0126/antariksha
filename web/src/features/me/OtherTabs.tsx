import { useEffect, useState } from 'react';
import { Button, Card, Checkbox, Chip, EmptyState, Field, Input, Notice, Select, Skeleton, useFeedback } from '../../ds';
import { LANGUAGES, useSettings, useT, type Lang } from '../../i18n';
import { api, downloadJSON } from '../../lib/api';
import { href } from '../../lib/router';
import { loadAccountCharts, localDataKeys, restoreDeviceCharts, saveAccountCharts, setProfileOwner, loadProfiles } from '../../ui/common/profiles';
import { LegacyImport } from '../../ui/common/LegacyImport';
import { PrivacyPage } from '../../ui/privacy/PrivacyPage';
import { AccountAccess } from './AccountAccess';

const signedOut = () => { setProfileOwner('signed-out'); window.dispatchEvent(new Event('antariksha-signed-out')); };

export function ChartsTab() {
  const { toast, confirm } = useFeedback();
  const [busy, setBusy] = useState(false);
  const [count, setCount] = useState(() => loadProfiles().profiles.length);
  async function run(fn: () => Promise<void> | void, ok: string) {
    setBusy(true);
    try { await fn(); toast(ok); setCount(loadProfiles().profiles.length); } catch (e) { toast(e instanceof Error ? e.message : 'Unable to sync charts.', 'danger'); } finally { setBusy(false); }
  }
  return (
    <div className="ds-stack">
      <Card title="Saved charts" sub={`${count} chart${count === 1 ? '' : 's'} on this device. Save them to your account to use them on another device. Only save charts you own or have permission to store.`}>
        <div className="ds-row">
          <Button variant="primary" busy={busy} onClick={() => void run(saveAccountCharts, 'Charts saved to your private account.')}>Save charts to my account</Button>
          <Button busy={busy} onClick={async () => { if (await confirm({ title: 'Load account charts?', body: 'Charts saved in your account replace the list on this device. A recovery copy of the current device charts is kept.', confirm: 'Load charts' })) void run(loadAccountCharts, 'Account charts loaded.'); }}>Load account charts</Button>
          <Button variant="ghost" busy={busy} onClick={async () => { if (await confirm({ title: 'Restore previous device charts?', body: 'Brings back the charts from before your last account load. The account copy does not change.', confirm: 'Restore' })) void run(restoreDeviceCharts, 'Previous device charts restored. Account copy unchanged.'); }}>Restore previous device charts</Button>
        </div>
        <a href={href('kundali')}>Open Kundali →</a>
      </Card>
      <Card><LegacyImport /></Card>
    </div>
  );
}

type Notice_ = { id: number; kind: 'follow' | 'family' | 'interest' | 'message' | 'accepted'; handle: string; created_at: string; read: boolean };
const LABEL: Record<Notice_['kind'], string> = { follow: 'sent you a follow request', family: 'invited you to a private family group', interest: 'sent you a matrimony interest', message: 'sent you a message', accepted: 'accepted your matrimony interest' };
const DEST: Record<Notice_['kind'], string> = { follow: href('community', 'people'), family: href('community', 'family'), interest: href('matrimony', 'interests'), message: href('matrimony', 'interests'), accepted: href('matrimony', 'interests') };

export function NotificationsTab() {
  const [items, setItems] = useState<Notice_[]>(), [more, setMore] = useState(false), [error, setError] = useState('');
  async function load(before?: number) {
    try { const rows = await api<Notice_[]>('/api/me/notifications' + (before ? `?before=${before}` : '')); setItems(old => before ? [...(old ?? []), ...rows] : rows); setMore(rows.length === 50); }
    catch (e) { setError(e instanceof Error ? e.message : 'Could not load'); }
  }
  useEffect(() => { void load(); }, []);
  async function update(id: number, action: 'read' | 'dismiss') {
    await api(`/api/me/notifications/${id}`, 'POST', { action });
    setItems(old => action === 'dismiss' ? old?.filter(n => n.id !== id) : old?.map(n => n.id === id ? { ...n, read: true } : n));
    window.dispatchEvent(new Event('astrisk-notifications'));
  }
  if (error) return <Notice tone="danger">{error}</Notice>;
  if (!items) return <Card><Skeleton /></Card>;
  return (
    <Card title="Notifications" sub="Invitations, interests and messages. Message text is never shown here." actions={<Button size="sm" variant="ghost" onClick={() => void load()}>Refresh notifications</Button>}>
      {items.length === 0 ? <EmptyState title="No notifications yet.">When someone sends you an interest, a message or an invitation, it appears here and on the bell.</EmptyState> : (
        <ul className="me-list">
          {items.map(n => (
            <li key={n.id} className={n.read ? '' : 'unread'}>
              <div><b>@{n.handle}</b> {LABEL[n.kind]} {!n.read && <Chip tone="accent">New</Chip>}<span className="ds-muted ds-small">{new Date(n.created_at).toLocaleString(undefined, { day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit' })}</span></div>
              <div className="ds-row">
                <a className="ds-btn ds-btn--secondary ds-btn--sm" href={DEST[n.kind]} onClick={() => void update(n.id, 'read')}>Open</a>
                {!n.read && <Button size="sm" variant="ghost" onClick={() => void update(n.id, 'read')}>Mark read</Button>}
                <Button size="sm" variant="ghost" onClick={() => void update(n.id, 'dismiss')}>Dismiss</Button>
              </div>
            </li>
          ))}
        </ul>
      )}
      {more && <Button variant="ghost" onClick={() => void load(items[items.length - 1].id)}>Older notifications</Button>}
    </Card>
  );
}

type Sec = { sessions: number; phone_verified: boolean; phone_available: boolean };
export function SecurityTab() {
  const { toast, confirm } = useFeedback();
  const [status, setStatus] = useState<Sec>(), [password, setPassword] = useState(''), [recovery, setRecovery] = useState('');
  const [phone, setPhone] = useState(''), [code, setCode] = useState(''), [consent, setConsent] = useState(false), [busy, setBusy] = useState(false);
  const [blocks, setBlocks] = useState<{ id: string; handle: string }[]>([]);
  const load = async () => { setStatus(await api<Sec>('/api/me/security')); setBlocks(await api('/api/community/blocks')); };
  useEffect(() => { void load().catch(e => toast(e.message, 'danger')); }, []);
  async function run(fn: () => Promise<void>) { setBusy(true); try { await fn(); } catch (e) { toast(e instanceof Error ? e.message : 'Request failed', 'danger'); } finally { setBusy(false); } }
  return (
    <div className="ds-stack">
      <Card title="Sign-in and recovery" sub={status ? `${status.sessions} active session${status.sessions === 1 ? '' : 's'}.` : undefined}>
        <Field label="Your current password" hint="Needed to create a recovery key or link a phone."><Input type="password" autoComplete="current-password" value={password} onChange={e => setPassword(e.target.value)} /></Field>
        <div className="ds-row">
          <Button busy={busy} disabled={!password} onClick={() => void run(async () => { const r = await api<{ recovery_key: string }>('/api/me/recovery-key', 'POST', { password }); setRecovery(r.recovery_key); setPassword(''); })}>Generate recovery key</Button>
          <Button variant="ghost" busy={busy} onClick={async () => { if (await confirm({ title: 'Sign out on all devices?', body: 'Every session ends, including this one.', confirm: 'Sign out everywhere' })) void run(async () => { await api('/api/me/logout-all', 'POST', {}); location.reload(); }); }}>Sign out on all devices</Button>
        </div>
        {recovery && <Notice tone="warning"><b>Your recovery key:</b> <code>{recovery}</code><br />Store it somewhere safe. It is shown once and replaces any earlier key; anyone with it and your handle can reset your password.</Notice>}
      </Card>
      <AccountAccess />
      <Card title="Mobile number" sub={status?.phone_verified ? 'Verified access to your number.' : 'Not verified.'}>
        {!status?.phone_available ? <p className="ds-muted">SMS verification starts once a text-message provider is set up.</p> : <>
          <form className="ds-row" onSubmit={e => { e.preventDefault(); void run(async () => { await api('/api/me/phone/start', 'POST', { phone, password, consent }); toast('Code sent. Enter it within 10 minutes.'); }); }}>
            <Field label="Phone in international format"><Input type="tel" placeholder="+919876543210" required value={phone} onChange={e => setPhone(e.target.value)} /></Field>
            <Checkbox label="Send a verification SMS to my number." required checked={consent} onChange={e => setConsent(e.target.checked)} />
            <Button type="submit" busy={busy}>Send code</Button>
          </form>
          <form className="ds-row" onSubmit={e => { e.preventDefault(); void run(async () => { await api('/api/me/phone/check', 'POST', { code }); setCode(''); await load(); toast('Number verified.'); }); }}>
            <Field label="Verification code"><Input inputMode="numeric" autoComplete="one-time-code" value={code} onChange={e => setCode(e.target.value)} /></Field>
            <Button type="submit" busy={busy}>Verify number</Button>
          </form></>}
      </Card>
      <Card title="Blocked accounts" sub="Unblocking does not restore earlier follows or matches.">
        {blocks.length === 0 ? <p className="ds-muted">No blocked accounts.</p> : <ul className="me-list">{blocks.map(b => <li key={b.id}><b>@{b.handle}</b><Button size="sm" variant="ghost" onClick={() => void run(async () => { await api('/api/community/blocks', 'POST', { target: b.id, block: false }); await load(); })}>Unblock</Button></li>)}</ul>}
      </Card>
    </div>
  );
}

export function PrivacyTab() {
  const { toast, confirm } = useFeedback();
  const [busy, setBusy] = useState(false);
  return (
    <div className="ds-stack">
      <Card title="Your data" sub="Download everything stored about you, or delete your account.">
        <div className="ds-row">
          <Button busy={busy} onClick={() => { setBusy(true); api('/api/me/export').then(d => downloadJSON('astrisk-account-data.json', d)).catch(e => toast(e.message, 'danger')).finally(() => setBusy(false)); }}>Export account data</Button>
          <Button variant="ghost" onClick={async () => { if (await confirm({ title: 'Sign out?', confirm: 'Sign out' })) { await api('/api/auth/logout', 'POST', {}); signedOut(); } }}>Sign out</Button>
        </div>
        <details>
          <summary>Delete this account</summary>
          <p>Permanently removes this account, posts, owned family groups, messages and conversations. Copies others saved outside the app cannot be removed.</p>
          <Button variant="danger" onClick={async () => {
            if (!await confirm({ title: 'Delete your account?', body: 'This permanently deletes your account and its data. It cannot be undone.', confirm: 'Delete my account', danger: true })) return;
            try { await api('/api/me', 'DELETE'); for (const k of localDataKeys()) localStorage.removeItem(k); signedOut(); } catch (e) { toast(e instanceof Error ? e.message : 'Could not delete', 'danger'); }
          }}>Delete my account</Button>
        </details>
      </Card>
      <PrivacyPage embedded />
    </div>
  );
}

export function SettingsTab() {
  const t = useT();
  const { lang, setLang, months, setMonths } = useSettings();
  return (
    <Card title={t('Settings')} sub="Saved on this device.">
      <div className="ds-form-grid">
        <Field label={t('Language')} hint="Menus, labels and astrology terms. Long explanations stay in English for now."><Select aria-label="Language" options={LANGUAGES} value={lang} onChange={e => setLang(e.target.value as Lang)} /></Field>
        <Field label="Month system" hint="Amanta (South and West India) or Purnimanta (North India) month names."><Select aria-label="Month system" options={[['amanta', 'Amanta'], ['purnimanta', 'Purnimanta']]} value={months} onChange={e => setMonths(e.target.value as 'amanta' | 'purnimanta')} /></Field>
      </div>
    </Card>
  );
}
