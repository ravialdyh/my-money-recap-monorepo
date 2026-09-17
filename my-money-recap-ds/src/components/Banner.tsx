import type { HTMLAttributes, ReactNode } from 'react';
import { Icon } from '../icons';
import { ensureStyles } from '../styles/ensureStyles';

export interface BannerProps extends HTMLAttributes<HTMLDivElement> {
  /** `warning` coral (Target Warning). `info` violet, for a neutral notice. `offline` grey, for
   *  "No connection. You can look, but not log." (PRD D58, D63). */
  tone?: 'warning' | 'info' | 'offline';
  /** Shows the dismiss control. The Target Warning is dismissable for the session only (PRD §6.8). */
  onDismiss?: () => void;
  children: ReactNode;
}
/** Full-width coral card that sits above the target card. Same radius as Card, no shadow. */
export function Banner({ tone = 'warning', onDismiss, children, className = '', ...rest }: BannerProps) {
  ensureStyles();
  return (
    <div className={`mmr-banner ${className}`} data-tone={tone} role="status" {...rest}>
      <div className="mmr-banner-body">{children}</div>
      {onDismiss ? (
        // The dismiss glyph is `plus` rotated 45 degrees: a stroked plus and a stroked cross are the
        // same shape, so the icon set stays at the twelve glyphs the PRD asks for.
        <button type="button" className="mmr-banner-dismiss mmr-pressable" aria-label="Dismiss" onClick={onDismiss}>
          <Icon name="plus" size={18} />
        </button>
      ) : null}
    </div>
  );
}
