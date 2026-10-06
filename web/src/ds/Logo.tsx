/** The Astrisk north star: four long rays for the directions and four short
 *  rays between them, optionally inside a thin orbit ring. */
const POINTS = '50,4 52.49,43.99 64.14,35.86 56.01,47.51 96,50 56.01,52.49 64.14,64.14 52.49,56.01 50,96 47.51,56.01 35.86,64.14 43.99,52.49 4,50 43.99,47.51 35.86,35.86 47.51,43.99';

export function NorthStar({ size = 32, ring = true, color = 'var(--c-accent)', ringColor = 'var(--c-accent-line)', title }: { size?: number; ring?: boolean; color?: string; ringColor?: string; title?: string }) {
  return (
    <svg width={size} height={size} viewBox="0 0 100 100" role={title ? 'img' : undefined} aria-label={title} aria-hidden={title ? undefined : true}>
      {ring && <circle cx="50" cy="50" r="47" fill="none" stroke={ringColor} strokeWidth={size < 40 ? 3 : 1.4} />}
      <polygon points={POINTS} fill={color} />
    </svg>
  );
}

export function Logo({ onClick, href = '#kundali', compact }: { onClick?: () => void; href?: string; compact?: boolean }) {
  return (
    <a className="ds-logo" href={href} onClick={onClick} aria-label="Astrisk home">
      <NorthStar size={compact ? 28 : 32} />
      <span className="ds-logo__word">ASTRISK</span>
    </a>
  );
}
