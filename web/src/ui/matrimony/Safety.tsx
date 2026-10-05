/** Practical safety guidance for matrimony, shown as its own tab. */
export function Safety() {
  const tips: [string, string][] = [
    ['Never send money', 'No genuine match asks for money, gift cards, crypto, customs fees or "help with a visa". Stop replying, block them and report the profile.'],
    ['Keep personal details private at first', 'Share your phone number, address, workplace or family details only after you trust someone. Use the in-app chat until then; your number is never shown unless you share it.'],
    ['Look for the verified badge', 'A verified badge means a moderator compared a live selfie with the profile photos. It does not check identity documents, income, education or marital status.'],
    ['Video call before meeting', 'A short video call confirms that the person matches their photos and helps you both feel comfortable.'],
    ['Meet in public, tell family', 'For first meetings choose a busy public place, arrange your own transport and tell a family member or friend where you are.'],
    ['Check what matters to you', 'Families often confirm education, work and marital status independently. Ask openly and respectfully; reluctance to answer basic questions is a signal.'],
    ['Dowry is illegal', 'Asking for or giving dowry is a crime in India under the Dowry Prohibition Act, 1961. Report any demand.'],
    ['Astrology is one input', 'A guna score or Mangal dosha is a traditional table, not a verdict on a person. Do not let a number override what you learn by talking.'],
  ];
  return (
    <section className="card mat-safety">
      <p className="kicker">Your safety comes first</p>
      <h2>Safe matchmaking</h2>
      <ol className="safety-list">{tips.map(([h, p]) => <li key={h}><b>{h}</b><p>{p}</p></li>)}</ol>
      <h3>Get help</h3>
      <ul>
        <li>Emergency: <b>112</b></li>
        <li>Women Helpline: <b>181</b></li>
        <li>Cyber crime and online fraud: <b>1930</b> or <a href="https://cybercrime.gov.in" target="_blank" rel="noopener noreferrer">cybercrime.gov.in</a></li>
      </ul>
      <p>Use <b>Report profile</b> on any card, or <b>Block</b> in a conversation. Moderators review every report; blocking immediately removes you from each other's lists, shortlists and shared contacts.</p>
    </section>
  );
}
