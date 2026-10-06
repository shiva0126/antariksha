import { useEffect, useRef, useState } from 'react';
import { FeedbackProvider } from './ds';
import { SettingsProvider } from './i18n';
import { api } from './lib/api';
import { useRoute, type Route } from './lib/router';
import { Shell, type Me } from './features/shell/Shell';
import { MePage } from './features/me/MePage';
import { AdminSection } from './features/admin/AdminSection';
import { PanchangSection } from './features/panchang/PanchangSection';
import { KundaliPage } from './ui/kundali/KundaliPage';
import { MatchPage } from './ui/match/MatchPage';
import { CommunityPage } from './ui/community/CommunityPage';
import { MatrimonyPage } from './ui/matrimony/MatrimonyPage';
import { BiodataPrint } from './ui/matrimony/BiodataPrint';
import { LoginPage } from './ui/LoginPage';
import { EmailLinkGate } from './ui/EmailRecovery';
import { ErrorBoundary } from './ui/common/ErrorBoundary';
import { initializeAccountCharts, setProfileOwner } from './ui/common/profiles';

function Page({ route, me }: { route: Route; me: Me }) {
  switch (route.section) {
    case 'kundali': return <KundaliPage sub={route.sub} />;
    case 'panchang': return <PanchangSection sub={route.sub} />;
    case 'matching': return <MatchPage />;
    case 'matrimony': return <MatrimonyPage />;
    case 'community': return <CommunityPage sub={route.sub} />;
    case 'me': return <MePage sub={route.sub} />;
    case 'admin': return <AdminSection sub={route.sub} role={me.role} />;
    case 'biodata': return <BiodataPrint />;
  }
}

function SignedIn({ me }: { me: Me }) {
  const route = useRoute();
  const signOut = async () => {
    await api('/api/auth/logout', 'POST', {}).catch(() => {});
    setProfileOwner('signed-out');
    window.dispatchEvent(new Event('antariksha-signed-out'));
  };
  return (
    <Shell route={route} me={me} onSignOut={() => void signOut()}>
      <ErrorBoundary key={route.section}><Page route={route} me={me} /></ErrorBoundary>
    </Shell>
  );
}

export default function App() {
  return <FeedbackProvider><EmailLinkGate><AuthenticatedApp /></EmailLinkGate></FeedbackProvider>;
}

/** Checks the session, binds saved charts to the account, and shows the
 *  sign-in screen when signed out. Re-checks when the window regains focus. */
function AuthenticatedApp() {
  const [me, setMe] = useState<Me | null>(null), [loading, setLoading] = useState(true);
  const sequence = useRef(0), identity = useRef('');
  const check = async () => {
    const run = ++sequence.current;
    const response = await fetch('/api/me', { credentials: 'same-origin' });
    if (run !== sequence.current) return;
    if (response.ok) {
      const m = await response.json();
      if (run !== sequence.current) return;
      if (identity.current !== m.id) { setLoading(true); setMe(null); }
      identity.current = m.id;
      setProfileOwner(m.id, { date: m.birth_date, time: m.birth_time?.slice(0, 5), place: m.birth_place ?? undefined });
      await initializeAccountCharts().catch(() => {});
      if (run !== sequence.current) return;
      setMe({ id: m.id, handle: m.handle, email: m.email, role: m.role || 'member', profile: m.profile });
    } else {
      identity.current = '';
      setProfileOwner('signed-out');
      setMe(null);
    }
    setLoading(false);
  };
  useEffect(() => {
    void check().catch(() => setLoading(false));
    const expired = () => { sequence.current++; identity.current = ''; setProfileOwner('signed-out'); setMe(null); setLoading(false); };
    const refresh = () => { void check().catch(() => { setMe(null); setLoading(false); }); };
    window.addEventListener('antariksha-signed-out', expired);
    window.addEventListener('focus', refresh);
    return () => { sequence.current++; window.removeEventListener('antariksha-signed-out', expired); window.removeEventListener('focus', refresh); };
  }, []);
  if (loading) return <main className="login-screen"><p>Opening Astrisk…</p></main>;
  if (!me) return <LoginPage onSignedIn={check} />;
  return <SettingsProvider key={me.id}><SignedIn me={me} /></SettingsProvider>;
}
