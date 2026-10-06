import { useEffect, useState } from 'react';
import { Button, Card, Checkbox, Field, FormGrid, Input, Notice, Select, Skeleton, StatTile, Textarea, useFeedback } from '../../ds';
import { api } from '../../lib/api';
import { longDate } from '../../astro/format';
import { PlaceSearch } from '../../ui/common/PlaceSearch';
import type { Place } from '../../ui/common/places';
import { MyCharacter } from '../../ui/community/MyCharacter';

type Member = { id: string; handle: string; email: string; email_verified: boolean; email_delivery_available: boolean; birth_date: string; birth_time: string; birth_place: { name: string; lat: number; lon: number; tz: string } | null; profile: Record<string, string> };
type Settings = { community: boolean; birth_date: string; avatar: string; accent: string; public_bio: string; interests: string[]; links: string[]; preferences: Record<string, string> };
const blankSettings: Settings = { community: false, birth_date: '', avatar: 'sun', accent: '#d6b467', public_bio: '', interests: [], links: [], preferences: {} };
const REFLECTIONS: [string, string][] = [['weekend', 'Your ideal weekend'], ['communication', 'How you prefer to communicate'], ['learning', 'Something you want to learn'], ['family', 'What family involvement means to you'], ['values', 'Values you want to live by']];
const SYMBOLS: [string, string][] = [['sun', 'Sun'], ['moon', 'Moon'], ['star', 'Star'], ['leaf', 'Leaf'], ['mountain', 'Mountain']];

function Basics({ me, onSaved }: { me: Member; onSaved: () => void }) {
  const { toast } = useFeedback();
  const [p, setP] = useState<Record<string, string>>(me.profile || {});
  const [place, setPlace] = useState<Place | undefined>(me.birth_place ? { name: me.birth_place.name, region: '', lat: me.birth_place.lat, lon: me.birth_place.lon, tz: me.birth_place.tz } : undefined);
  const [busy, setBusy] = useState(false), [msg, setMsg] = useState('');
  const set = (k: string, v: string) => setP(x => ({ ...x, [k]: v }));
  async function save(e: React.FormEvent) {
    e.preventDefault(); setBusy(true); setMsg('');
    try { await api('/api/me', 'PUT', { name: p.name ?? '', city: p.city ?? '', bio: p.bio ?? '', hobbies: p.hobbies ?? '', goals: p.goals ?? '' }); setMsg('Private profile saved.'); onSaved(); }
    catch (err) { toast(err instanceof Error ? err.message : 'Could not save', 'danger'); } finally { setBusy(false); }
  }
  async function savePlace(pl: Place) {
    try { await api('/api/me/birth-place', 'PUT', { name: [pl.name, pl.region].filter(Boolean).join(', '), lat: pl.lat, lon: pl.lon, tz: pl.tz }); setPlace(pl); toast('Birthplace saved.'); }
    catch (err) { toast(err instanceof Error ? err.message : 'Could not save birthplace', 'danger'); }
  }
  return (
    <Card title="Basics" sub={<>Signed in as <b>{me.email || '@' + me.handle}</b> · private to you</>}>
      <div className="ds-stats">
        <StatTile label="Birth date" value={me.birth_date ? longDate(me.birth_date) : 'Not recorded'} />
        <StatTile label="Birth time" value={me.birth_time?.slice(0, 5) || 'Not recorded'} />
        <StatTile label="Handle" value={/^member_[0-9a-f]{12}$/.test(me.handle) ? 'Not chosen' : '@' + me.handle} note={/^member_[0-9a-f]{12}$/.test(me.handle) ? 'Members see your display name' : undefined} />
        <StatTile label="Email" value={me.email_verified ? 'Verified' : 'Not verified'} note={!me.email_delivery_available ? 'Verification starts once email is set up' : undefined} />
      </div>
      {!me.email_verified && me.email_delivery_available && <div><Button size="sm" onClick={() => api<{ message: string }>('/api/me/email/verification', 'POST', {}).then(r => toast(r.message)).catch(e => toast(e.message, 'danger'))}>Send verification email</Button></div>}
      <PlaceSearch label="Birthplace (for lagna and Mangal dosha)" value={place} onChange={pl => void savePlace(pl)} />
      <form onSubmit={save} className="ds-stack">
        <FormGrid>
          <Field label="Display name"><Input maxLength={100} value={p.name ?? ''} onChange={e => set('name', e.target.value)} autoComplete="name" /></Field>
          <Field label="City"><Input maxLength={100} value={p.city ?? ''} onChange={e => set('city', e.target.value)} /></Field>
          <Field label="About me" span><Textarea maxLength={2000} value={p.bio ?? ''} onChange={e => set('bio', e.target.value)} /></Field>
          <Field label="Hobbies and interests"><Textarea maxLength={1000} rows={2} value={p.hobbies ?? ''} onChange={e => set('hobbies', e.target.value)} /></Field>
          <Field label="What I want to learn or experience"><Textarea maxLength={1000} rows={2} value={p.goals ?? ''} onChange={e => set('goals', e.target.value)} /></Field>
        </FormGrid>
        <div className="ds-row"><Button type="submit" variant="primary" busy={busy}>Save private profile</Button>{msg && <span role="status" className="ds-muted">{msg}</span>}</div>
      </form>
    </Card>
  );
}

function CommunityProfile() {
  const { toast } = useFeedback();
  const [s, setS] = useState<Settings>(), [interests, setInterests] = useState(''), [links, setLinks] = useState(''), [busy, setBusy] = useState(false);
  useEffect(() => {
    api<Settings[]>('/api/community/settings').then(rows => { const r = { ...blankSettings, ...rows[0] }; setS(r); setInterests(r.interests.join(', ')); setLinks(r.links.join('\n')); }).catch(e => toast(e.message, 'danger'));
  }, []);
  if (!s) return <Card title="Community profile"><Skeleton /></Card>;
  const set = (x: Partial<Settings>) => setS({ ...s, ...x });
  async function save(e: React.FormEvent) {
    e.preventDefault(); setBusy(true);
    try {
      await api('/api/community/settings', 'PUT', { ...s, interests: [...new Set(interests.split(',').map(x => x.trim()).filter(Boolean))], links: links.split('\n').map(x => x.trim()).filter(Boolean) });
      toast('Profile and privacy choices saved.');
    } catch (err) { toast(err instanceof Error ? err.message : 'Could not save', 'danger'); } finally { setBusy(false); }
  }
  return (
    <Card title="Community profile" sub="What other members see when you join the community. Your birth date and reflections stay private.">
      <form onSubmit={save} className="ds-stack">
        <Checkbox label="Join the adult community and share my community introduction, avatar, interests and links." checked={s.community} onChange={e => set({ community: e.target.checked })} />
        <FormGrid>
          <Field label="Avatar symbol"><Select options={SYMBOLS} value={s.avatar} onChange={e => set({ avatar: e.target.value })} /></Field>
          <Field label="Accent colour"><input type="color" value={s.accent} onChange={e => set({ accent: e.target.value })} /></Field>
          <Field label="Private birth date" hint="Used only to confirm you are an adult."><Input required type="date" value={s.birth_date || ''} onChange={e => set({ birth_date: e.target.value })} /></Field>
          <Field label="Interests, separated by commas"><Input maxLength={800} value={interests} onChange={e => setInterests(e.target.value)} placeholder="Hiking, books, cooking" /></Field>
          <Field label="Community introduction" span><Textarea maxLength={1000} value={s.public_bio} onChange={e => set({ public_bio: e.target.value })} /></Field>
          <Field label="Social profile links" hint="One https link per line. Self-declared; never verified or searched." span><Textarea rows={2} value={links} onChange={e => setLinks(e.target.value)} /></Field>
        </FormGrid>
        <details>
          <summary className="ds-muted">Private reflections (only you see these)</summary>
          <FormGrid>{REFLECTIONS.map(([k, l]) => <Field key={k} label={l}><Textarea maxLength={200} rows={2} value={s.preferences[k] || ''} onChange={e => set({ preferences: { ...s.preferences, [k]: e.target.value } })} /></Field>)}</FormGrid>
        </details>
        <div><Button type="submit" variant="primary" busy={busy}>Save profile & privacy</Button></div>
      </form>
    </Card>
  );
}

export function ProfileTab() {
  const [me, setMe] = useState<Member>(), [error, setError] = useState('');
  const load = () => api<Member>('/api/me').then(setMe).catch(e => setError(e.message));
  useEffect(() => { void load(); }, []);
  if (error) return <Notice tone="danger">{error}</Notice>;
  if (!me) return <Card><Skeleton lines={5} /></Card>;
  return (
    <div className="ds-stack">
      <Basics me={me} onSaved={() => void load()} />
      <CommunityProfile />
      <section aria-label="My character" className="ds-stack">
        <h2 className="ds-card__title">My character & interests</h2>
        <MyCharacter />
      </section>
    </div>
  );
}
