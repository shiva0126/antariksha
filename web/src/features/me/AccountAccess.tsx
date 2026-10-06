import { useEffect, useState } from 'react';
import { Button, Card, Field, Input, Notice, useFeedback } from '../../ds';
import { api } from '../../lib/api';

type Session = { id: string; client_label: string; created_at: string | null; expires_at: string; current: boolean };
type Security = { email_verified: boolean; email_delivery_available: boolean };

export function AccountAccess() {
  const { toast, confirm } = useFeedback();
  const [password,setPassword] = useState(''), [next,setNext] = useState(''), [repeat,setRepeat] = useState('');
  const [sessions,setSessions] = useState<Session[]>([]), [security,setSecurity] = useState<Security>();
  const [busy,setBusy] = useState(false), [error,setError] = useState('');
  async function load() { const [rows, status] = await Promise.all([api<Session[]>('/api/me/sessions'),api<Security>('/api/me/security')]);setSessions(rows);setSecurity(status); }
  useEffect(() => { void load().catch(e=>setError(e.message)); },[]);
  return <>
    <Card title="Change password" sub="Changing your password signs out every device, including this one, and invalidates old recovery keys and reset links.">
      <form className="ds-stack" onSubmit={async e=>{e.preventDefault();setError('');if(next!==repeat){setError('New passwords do not match.');return;}setBusy(true);try{const result=await api<{message:string}>('/api/me/password','POST',{password,new_password:next});setPassword('');setNext('');setRepeat('');toast(result.message);window.dispatchEvent(new Event('antariksha-signed-out'));}catch(e){setError(e instanceof Error?e.message:'Password change failed');}finally{setBusy(false);}}}>
        <Field label="Current password for account changes"><Input required type="password" autoComplete="current-password" maxLength={72} value={password} onChange={e=>setPassword(e.target.value)}/></Field>
        <Field label="New password" hint="12–72 bytes. Use a unique password saved in your password manager."><Input required type="password" autoComplete="new-password" minLength={12} maxLength={72} value={next} onChange={e=>setNext(e.target.value)}/></Field>
        <Field label="Confirm new password"><Input required type="password" autoComplete="new-password" value={repeat} onChange={e=>setRepeat(e.target.value)}/></Field>
        {error&&<Notice tone="danger">{error}</Notice>}
        <Button type="submit" busy={busy}>Change password and sign out</Button>
      </form>
    </Card>
    <Card title="Active sessions" sub="Browser/device labels are approximate, not verified device identities. No IP addresses or precise locations are stored here.">
      <p>To revoke a session, enter your current password in the account changes field above. You do not need to enter a new password.</p>
      <ul className="me-list">{sessions.map(session=><li key={session.id}><div><b>{session.client_label}{session.current?' · This session':''}</b><p>{session.created_at?'Signed in '+new Date(session.created_at).toLocaleString():'Sign-in date unavailable for this older session'} · Expires {new Date(session.expires_at).toLocaleString()}</p></div><Button disabled={!password} busy={busy} onClick={async()=>{if(!await confirm({title:session.current?'Sign out this session?':'Revoke this session?',body:'That session will no longer be able to access your account.',confirm:'Revoke session'}))return;setBusy(true);try{const result=await api<{signed_out:boolean}>(`/api/me/sessions/${session.id}/revoke`,'POST',{password});setPassword('');if(result.signed_out)window.dispatchEvent(new Event('antariksha-signed-out'));else{await load();toast('Session revoked.');}}catch(e){toast(e instanceof Error?e.message:'Could not revoke session','danger');}finally{setBusy(false);}}}>Revoke session</Button></li>)}</ul>
    </Card>
    <Card title="Email verification" sub={security?.email_verified?'Your email is verified.':'Your email has not been verified.'}>
      {security&&!security.email_delivery_available?<Notice tone="warning">Email delivery is not configured yet. Save a recovery key using Sign-in and recovery; email password resets are currently unavailable.</Notice>:security&&!security.email_verified&&<Button busy={busy} onClick={async()=>{setBusy(true);try{await api('/api/me/email/verification','POST',{});toast('If eligible, a verification link will be sent. Check your inbox.');}catch(e){toast(e instanceof Error?e.message:'Could not request verification','danger');}finally{setBusy(false);}}}>Send verification email</Button>}
    </Card>
  </>;
}
