import { useState } from 'react';
import { useT } from '../../i18n';

/** Renders an Ask Astrisk answer. New answers have three parts under fixed
 *  headings ("In short", "What this means for you", "Chart details"); the
 *  technical details start folded. Older answers are shown as paragraphs.
 *  The headings stay English in the text as markers and are shown in the
 *  member's language. A machine-translated answer carries its English
 *  original, which the member can switch to. */
const SHORT = 'In short\n';
const POINTS = 'What this means for you\n';
const DETAILS = 'Chart details\n';
const CLOSING = 'For reflection, not certainty.';

export function AnswerText({ text, original }: { text: string; original?: string }) {
  const t = useT();
  const [english, setEnglish] = useState(false);
  if (original) {
    return (
      <>
        <Parts text={english ? original : text} t={english ? (s: string) => s : t} />
        <p className="answer-translated muted small">
          {t('Machine translation from English')} · <button type="button" className="linklike" onClick={() => setEnglish(v => !v)}>{english ? t('Show translation') : t('Show English')}</button>
        </p>
      </>
    );
  }
  return <Parts text={text} t={t} />;
}

function Parts({ text, t }: { text: string; t: (s: string) => string }) {
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
      <p className="answer-label">{t('In short')}</p>
      <p className="answer-short">{short}</p>
      {points.length > 0 && <>
        <p className="answer-label">{t('What this means for you')}</p>
        <ul className="answer-points">{points.map((x, i) => <li key={i}>{x}</li>)}</ul>
      </>}
      {details.length > 0 && (
        <details className="answer-details">
          <summary>{t('Chart details')}</summary>
          {details.map((d, i) => <p key={i}>{d}</p>)}
        </details>
      )}
      {closing && <p className="answer-closing">{t(CLOSING)}</p>}
    </div>
  );
}
