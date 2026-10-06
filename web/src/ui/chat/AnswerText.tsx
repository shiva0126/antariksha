/** Renders an Ask Astrisk answer. New answers have three parts under fixed
 *  headings ("In short", "What this means for you", "Chart details"); the
 *  technical details start folded. Older answers are shown as paragraphs. */
const SHORT = 'In short\n';
const POINTS = 'What this means for you\n';
const DETAILS = 'Chart details\n';
const CLOSING = 'For reflection, not certainty.';

export function AnswerText({ text }: { text: string }) {
  if (!text.startsWith(SHORT)) return <>{text.split(/\n\n+/).map((p, i) => <p key={i}>{p}</p>)}</>;
  const blocks = text.split(/\n\n+/);
  let short = '', points: string[] = [], details: string[] = [], closing = false;
  let inDetails = false;
  for (const b of blocks) {
    if (b === CLOSING) { closing = true; continue; }
    if (b.startsWith(SHORT)) short = b.slice(SHORT.length);
    else if (b.startsWith(POINTS)) points = b.slice(POINTS.length).split('\n').map(x => x.replace(/^•\s*/, ''));
    else if (b.startsWith(DETAILS)) { inDetails = true; details.push(b.slice(DETAILS.length)); }
    else if (inDetails) details.push(b);
    else short += (short ? ' ' : '') + b;
  }
  return (
    <div className="answer">
      <p className="answer-label">In short</p>
      <p className="answer-short">{short}</p>
      {points.length > 0 && <>
        <p className="answer-label">What this means for you</p>
        <ul className="answer-points">{points.map((x, i) => <li key={i}>{x}</li>)}</ul>
      </>}
      {details.length > 0 && (
        <details className="answer-details">
          <summary>Chart details</summary>
          {details.map((d, i) => <p key={i}>{d}</p>)}
        </details>
      )}
      {closing && <p className="answer-closing">{CLOSING}</p>}
    </div>
  );
}
