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
      for (const k of localDataKeys().filter(k => k.startsWith('antariksha.chat.'))) { const sid = localStorage.getItem(k); if (sid) sids.add(sid); }
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
        <p>Antariksha is free. It takes no payment, has no premium tier and sells no remedies, gemstones or consultations.</p>
        <h3>What is stored, and where</h3>
        <p><b>On this device:</b> the birth profiles you save, your language and calendar settings, and a random ID linking each profile to its conversation.</p>
        <p><b>On the server:</b> when you use Ask Antariksha, your questions and the answers are stored with the birth details of that chart, so your history is there when you return. Charts, Panchang, matching and muhurta results are calculated on request. Readings are cached by chart and date without any name. Nothing is shared with third parties. If a language model is enabled, the question and the chart facts are sent to it to write the answer.</p>
        <h3>Purpose and consent</h3>
        <p>Birth details are used only to calculate charts and answer your questions. You agree to this when you create a chart, and you can withdraw at any time by deleting your data. Users under 18 need a parent's or guardian's permission.</p>
        <h3>Delete your data</h3>
        <p>Deleting a profile (Kundali → Delete) removes its conversation from the server. The button below deletes everything this device has stored, and every conversation linked from it.</p>
        <button className="primary danger-bg" onClick={deleteAll}>Delete all my data</button>
        {status && <p role="status" className="good-text">{status}</p>}
        <h3>Credits</h3>
        <p className="muted small">Astronomy: Swiss Ephemeris (Astrodienst, AGPL). Place data © GeoNames (geonames.org), CC BY 4.0. Classical passages: The Brihat Jataka of Varaha Mihira, tr. N. Chidambaram Iyer (1885), public domain.</p>
      </article>
    </div>
  );
}
