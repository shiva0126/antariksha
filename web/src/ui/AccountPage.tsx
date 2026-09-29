import { useEffect, useState } from 'react';

type Profile = { name: string; city: string; bio: string; hobbies: string; goals: string };
type Member = { handle: string; profile: Partial<Profile> };
const empty: Profile = { name: '', city: '', bio: '', hobbies: '', goals: '' };
async function call(path: string, method = 'GET', body?: unknown) {
  const response = await fetch(path, { method, credentials: 'same-origin', headers: { 'Content-Type': 'application/json' }, body: body === undefined ? undefined : JSON.stringify(body) });
  const result = await response.json();
  if (!response.ok) throw new Error(result.error || 'Request failed');
  return result;
}
export function AccountPage() {
  const [member, setMember] = useState<Member | null>(null);
  const [profile, setProfile] = useState<Profile>(empty);
  const [handle, setHandle] = useState('');
  const [password, setPassword] = useState('');
  const [register, setRegister] = useState(false);
  const [consent, setConsent] = useState(false);
  const [busy, setBusy] = useState(false);
  const [ready, setReady] = useState(false);
  const [message, setMessage] = useState('');
  const load = async () => { const value = await call('/api/me'); setMember(value); setProfile({ ...empty, ...value.profile }); };
  useEffect(() => { load().catch(() => {}).finally(() => setReady(true)); }, []);
  async function action(fn: () => Promise<void>) { setBusy(true); setMessage(''); try { await fn(); } catch (error) { setMessage(error instanceof Error ? error.message : 'Request failed'); } finally { setBusy(false); } }
  return <section style={{ maxWidth: 680, margin: '3rem auto', padding: '0 1.5rem' }}>
    <p>YOUR PRIVATE SPACE</p><h1>My Antariksha</h1>
    <p>Keep your interests, goals and introduction together. Your profile is private and is not listed for matchmaking.</p>
    {!ready ? <p>Loading account…</p> : !member ? <form onSubmit={event => { event.preventDefault(); void action(async () => { await call('/api/auth/' + (register ? 'register' : 'login'), 'POST', { handle, password, consent }); setPassword(''); await load(); }); }}>
      <h2>{register ? 'Create an account' : 'Welcome back'}</h2>
      <p>Phone sign-in is coming after SMS delivery is configured. You can use a handle and password now.</p>
      <label style={{ display: 'grid', gap: 8, marginBottom: 16 }}>Handle<input required autoComplete="username" pattern="[a-z0-9_]{3,32}" value={handle} onChange={e => setHandle(e.target.value)} /></label>
      <label style={{ display: 'grid', gap: 8, marginBottom: 16 }}>Password<input required type="password" minLength={12} maxLength={72} autoComplete={register ? 'new-password' : 'current-password'} value={password} onChange={e => setPassword(e.target.value)} /></label>
      <p>Use at least 12 characters. Keep your password safe; account recovery is not available yet.</p>
      {register && <label><input type="checkbox" required checked={consent} onChange={e => setConsent(e.target.checked)} /> I agree to storing this private account and profile. I can export or delete it here.</label>}
      <p><button disabled={busy} type="submit">{busy ? 'Please wait…' : register ? 'Create account' : 'Sign in'}</button> <button type="button" onClick={() => { setRegister(!register); setMessage(''); }}>{register ? 'Already have an account?' : 'Create account'}</button></p>
    </form> : <>
      <p>Signed in as <strong>@{member.handle}</strong> · Private</p>
      <form onSubmit={event => { event.preventDefault(); void action(async () => { await call('/api/me', 'PUT', profile); setMessage('Private profile saved.'); }); }}>
        {(Object.keys(empty) as (keyof Profile)[]).map(key => <label key={key} style={{ display: 'grid', gap: 8, marginBottom: 16 }}>{({ name: 'Display name', city: 'City', bio: 'About me', hobbies: 'Hobbies and interests', goals: 'What I want to learn or experience' })[key]}
          <textarea rows={key === 'name' || key === 'city' ? 1 : 3} maxLength={key === 'name' || key === 'city' ? 100 : key === 'bio' ? 2000 : 1000} value={profile[key]} onChange={e => setProfile({ ...profile, [key]: e.target.value })} />
        </label>)}
        <button disabled={busy}>Save private profile</button>
      </form>
      <p><button disabled={busy} onClick={() => void action(async () => { const data = await call('/api/me'); const url = URL.createObjectURL(new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })); const link = document.createElement('a'); link.href = url; link.download = 'antariksha-profile.json'; link.click(); URL.revokeObjectURL(url); })}>Export profile</button> <button disabled={busy} onClick={() => void action(async () => { await call('/api/auth/logout', 'POST', {}); setMember(null); setProfile(empty); })}>Sign out</button></p>
      <details><summary>Delete this account</summary><p>This deletes the server account, private profile and all its sessions. Existing browser-saved charts and astrology conversations are separate.</p><button disabled={busy} onClick={() => { if (window.confirm('Permanently delete this account and private profile?')) void action(async () => { await call('/api/me', 'DELETE'); setMember(null); setProfile(empty); setMessage('Account deleted.'); }); }}>Delete my account</button></details>
    </>}
    <p role="status" aria-live="polite">{message}</p>
  </section>;
}
