import { useEffect } from 'react';
import { MMR_CSS } from './css.generated';

const STYLE_ID = 'mmr-design-system-styles';

/** Injects the design-system stylesheet once per document. Safe to call from every component. */
export function ensureStyles(doc: Document | undefined = typeof document !== 'undefined' ? document : undefined): void {
  if (!doc || doc.getElementById(STYLE_ID)) return;
  const el = doc.createElement('style');
  el.id = STYLE_ID;
  el.textContent = MMR_CSS;
  doc.head.appendChild(el);
}

export function useStyles(): void {
  // Inject synchronously during render for SSR-less previews, then again on mount for safety.
  ensureStyles();
  useEffect(() => { ensureStyles(); }, []);
}
