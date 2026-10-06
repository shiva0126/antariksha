import { useEffect,useState } from 'react';
import { EmailRecovery } from './EmailRecovery';
import { NorthStar } from '../ds';
import { PlaceSearch } from './common/PlaceSearch';
import type { Place } from './common/places';
import { useRoute } from '../lib/router';
import './login.css';

export function LoginPage({ onSignedIn }: { onSignedIn: () => Promise<void> }) {
  const adminEntry = useRoute().section === 'admin';
  const [emailRecovery,setEmailRecovery]=useState(false),[emailAvailable,setEmailAvailable]=useState(false);
  useEffect(()=>{fetch('/api/auth/options').then(r=>r.json()).then(o=>setEmailAvailable(o.email_delivery_available)).catch(()=>{});},[]);
  const [mode, setMode] = useState<'login' | 'register' | 'recover'>('login');
  const [email, setEmail] = useState(''), [password, setPassword] = useState('');
  const [date, setDate] = useState(''), [time, setTime] = useState(''), [key, setKey] = useState('');
  const [name, setName] = useState(''), [handle, setHandle] = useState(''), [place, setPlace] = useState<Place>();
  // Sign-up is two short steps: who you are, then when and where you were born.
  const [step, setStep] = useState<1 | 2>(1);
  const [busy, setBusy] = useState(false), [error, setError] = useState(''), [legacy, setLegacy] = useState(false);
  const title = mode === 'register' ? 'Begin your journey' : mode === 'recover' ? 'Recover your account' : adminEntry ? 'Administrator sign in' : 'Welcome to Astrisk';
  if(emailRecovery)return <EmailRecovery action="request" onBack={()=>setEmailRecovery(false)}/>;
  return <main className="login-screen"><section className="login-card">
    <div className="login-emblem" aria-hidden><NorthStar size={56} /></div><p className="kicker">ASTRISK.SPACE</p>
    <h1>{title}</h1><p className="login-intro">{mode === 'register' ? 'A few details to create your private account.' : mode === 'recover' ? 'Use the recovery key you saved in account security.' : adminEntry ? 'Use the email and password of your existing administrator account. Access is checked securely after sign-in.' : 'Sign in to enter your personal universe.'}</p>
    {adminEntry && mode === 'login' && <p>No separate admin password is needed. Creating a new account does not grant administrator access.</p>}
    <form onSubmit={async event => { event.preventDefault(); if (mode === 'register' && step === 1) { setError(''); setStep(2); return; } setBusy(true); setError(''); try {
      const response = await fetch('/api/auth/' + mode, { method: 'POST', credentials: 'same-origin', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(mode === 'recover' ? { handle: email, key, password } : { ...(legacy ? { handle: email } : { email }), password, ...(mode === 'register' ? { birth_date: date, birth_time: time, consent: true, name, ...(handle ? { handle: handle.toLowerCase() } : {}), ...(place ? { birth_place: { name: [place.name, place.region].filter(Boolean).join(', '), lat: place.lat, lon: place.lon, tz: place.tz } } : {}) } : {}) }) });
      const result = await response.json(); if (!response.ok) throw new Error(result.error || 'Unable to sign in');
      if (mode === 'recover') { setMode('login'); setPassword(''); setKey(''); setError('Password reset. Sign in with your new password.'); }
      else { setPassword(''); await onSignedIn(); }
    } catch (e) { setError(e instanceof Error ? e.message : 'Unable to connect. Please try again.'); } finally { setBusy(false); } }}>
      {!(mode === 'register' && step === 2) && <><label className="field"><span>{legacy ? 'Existing handle' : 'Email'}</span><input required type={legacy ? 'text' : 'email'} autoComplete="username" value={email} maxLength={254} onChange={e => setEmail(e.target.value)} placeholder={legacy ? 'Your existing handle' : 'you@example.com'} /></label>
      <label className="field"><span>{mode === 'recover' ? 'New password' : 'Password'}</span><input required type="password" minLength={mode === 'login' ? undefined : 12} maxLength={72} autoComplete={mode === 'login' ? 'current-password' : 'new-password'} value={password} onChange={e => setPassword(e.target.value)} placeholder={mode === 'login' ? 'Your password' : 'At least 12 characters'} /></label></>}
      {mode === 'register' && <p className="login-step">Step {step} of 2 · {step === 1 ? 'About you' : 'Your birth details'}</p>}
      {mode === 'register' && step === 1 && <><label className="field"><span>Your name</span><input required maxLength={100} autoComplete="name" value={name} onChange={e => setName(e.target.value)} placeholder="How others will see you" /></label>
        <label className="field"><span>Handle <em>(optional)</em></span><input maxLength={32} pattern="[a-zA-Z0-9_]{3,32}" value={handle} onChange={e => setHandle(e.target.value)} placeholder="e.g. ananya_k — letters, numbers, _" /></label></>}
      {mode === 'register' && step === 2 && <div className="login-birth"><label className="field"><span>Date of birth</span><input required type="date" min="1900-01-01" max={new Date().toISOString().slice(0, 10)} value={date} onChange={e => setDate(e.target.value)} /></label><label className="field"><span>Birth time</span><input required type="time" value={time} onChange={e => setTime(e.target.value)} /></label></div>}
      {mode === 'register' && step === 2 && <PlaceSearch label="Birthplace" value={place} onChange={setPlace} />}
      {mode === 'recover' && <label className="field"><span>Recovery key</span><input required type="password" autoComplete="off" value={key} onChange={e => setKey(e.target.value)} /></label>}
      {mode === 'register' && step === 2 && <p className="login-note">Enter your accurate birth details. Creating an account agrees to private storage of these details; you can export or delete them in your account. Interests and hobbies come later.</p>}
      <p role="status" aria-live="polite" className="login-error">{error}</p>
      <button className="ds-btn ds-btn--primary ds-btn--block" disabled={busy}>{busy ? 'Please wait…' : mode === 'register' ? (step === 1 ? 'Continue' : 'Create account') : mode === 'recover' ? 'Reset password' : 'Sign in'}</button>
    </form>
    <div className="login-links">{mode === 'register' && step === 2 && <button onClick={() => setStep(1)}>← Back</button>}<button onClick={() => { setMode(mode === 'register' ? 'login' : 'register'); setStep(1); setLegacy(false); setError(''); }}>{mode === 'register' ? 'Already registered? Sign in' : 'New here? Create an account'}</button>{mode === 'login' && <button onClick={() => { setMode('recover'); setError(''); }}>Forgot password?</button>}</div>
    {mode==='recover'&&<p>{emailAvailable?<button onClick={()=>setEmailRecovery(true)}>Send an email reset link instead</button>:'Email reset is not available yet. Use your saved recovery key.'}</p>}
    {mode === 'login' && <details className="login-legacy"><summary>Existing account without an email?</summary><label><input type="checkbox" checked={legacy} onChange={e => setLegacy(e.target.checked)} /> Use my original handle</label></details>}
    <p className="login-foot">Your profile starts private. You choose what to share.</p>
    {mode === 'login' && <a href={adminEntry ? '#kundali' : '#admin'}>{adminEntry ? 'Back to member sign in' : 'Administrator sign in'}</a>}
  </section></main>;
}
