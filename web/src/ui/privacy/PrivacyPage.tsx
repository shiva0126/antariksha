import { useState } from 'react';
import { deleteChatSession } from '../../api/client';
import { chatKey, loadProfiles, localDataKeys } from '../common/profiles';

export function PrivacyPage() {
  const [status, setStatus] = useState('');
  async function deleteAll() {
    const { profiles } = loadProfiles();
    // Every session linked from this device: each saved profile, plus charts
    // that were chatted about but are no longer saved. Each is deleted once.
    const sids = new Set<string>();
    try {
      for (const p of profiles) { const sid = localStorage.getItem(chatKey(p)); if (sid) sids.add(sid); }
      for (const k of localDataKeys().filter(k => k.includes('.chat.'))) { const sid = localStorage.getItem(k); if (sid) sids.add(sid); }
    } catch { /* storage unavailable */ }
    let removed = 0;
    for (const sid of sids) {
      try { await deleteChatSession(sid); removed++; } catch { /* already gone */ }
    }
    for (const k of localDataKeys()) { try { localStorage.removeItem(k); } catch { /* ignore */ } }
    try { sessionStorage.clear(); } catch { /* ignore */ }
    setStatus(`Deleted ${removed} conversation${removed === 1 ? '' : 's'} from the server and all profiles and settings from this device.`);
  }
  return (
    <div className="page privacy">
      <header className="page-head"><div><p className="kicker">Your data</p><h1>Privacy</h1></div></header>
      <article className="card prose">
        <h3>Free, with no payments</h3>
        <p>Astrisk is free. It takes no payment, has no premium tier and sells no remedies, gemstones or consultations.</p>
        <h3>What is stored, and where</h3>
        <p><b>On this device:</b> the birth profiles you save, your language and calendar settings, and a random ID linking each profile to its conversation.</p>
        <p><b>Account chart storage:</b> “Save charts to my account” stores a private copy on the server for other devices. Existing device charts are uploaded only when you choose to save. Deleting a device chart does not remove the account copy until you save the changed list. Loading account charts preserves a recovery copy on this device. Account deletion removes the server copy.</p>
        <p><b>On the server:</b> your email, password hash, birth details, profiles, posts, connections, family groups and messages are stored. Conversations retain the chart details and questions; readings are cached. If a language model is enabled, questions and chart facts are sent to that provider. Phone verification, when configured, sends your number to the SMS provider; the number is stored encrypted. Community posts follow their selected audience. Social links are supplied by you, not discovered by searching for your identity.</p>
        <h3>Purpose and consent</h3>
        <p>Birth details support charts and age eligibility. Community and matrimony require adult opt-in. Hobbies and introductions are optional. Family information should only be added with permission; avoid sensitive information about children. Sharing with other members allows them to save copies outside this service.</p>
        <h3>Delete your data</h3>
        <p>Deleting a chart profile removes its conversation. The button below clears this account's charts on this device and linked conversations, not the account or community data. For full export visit Community → Security. To delete the server account and its data visit <a href="#account">My account</a>.</p>
        <button className="primary danger-bg" onClick={deleteAll}>Delete charts on this device</button>
        {status && <p role="status" className="good-text">{status}</p>}
        <h3>Credits</h3>
        <p className="muted small">Astronomy: Swiss Ephemeris (Astrodienst, AGPL). Place data © GeoNames (geonames.org), CC BY 4.0. Classical passages: The Brihat Jataka of Varaha Mihira, tr. N. Chidambaram Iyer (1885), public domain.</p>
      </article>
    </div>
  );
}
