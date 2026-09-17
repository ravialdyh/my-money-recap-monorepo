import { MMR_CSS } from './css.generated';

const STYLE_ID = 'mmr-design-system-styles';

/** Injects the design-system stylesheet once per document. Every component calls it during render so the
 * stylesheet is present before the first paint; a no-op on the server. */
export function ensureStyles(): void {
  if (typeof document === 'undefined' || document.getElementById(STYLE_ID)) return;
  const el = document.createElement('style');
  el.id = STYLE_ID;
  el.textContent = MMR_CSS;
  document.head.appendChild(el);
}
