import type { HTMLAttributes, ReactNode } from 'react';
import { useStyles } from '../styles/useStyles';

export interface SheetProps extends HTMLAttributes<HTMLDivElement> {
  /** Sheet title, rendered in Fredoka beside the handle. */
  title?: string;
  /** `bottom` a sheet over the current screen · `full` a white surface filling the ground (Settings). */
  variant?: 'bottom' | 'full';
  /** Draws the dimming layer above the screen behind it. */
  scrim?: boolean;
  children?: ReactNode;
}
/** White surface with `radius.card` on the top corners and a drag handle. The one surface that is not a ground. */
export function Sheet({ title, variant = 'bottom', scrim = false, children, className = '', ...rest }: SheetProps) {
  useStyles();
  const sheet = (
    <div className={`mmr-sheet ${className}`} data-variant={variant} role="dialog" aria-label={title} {...rest}>
      {variant === 'bottom' ? <span className="mmr-sheet-handle" /> : null}
      {title ? <h2 className="mmr-sheet-title">{title}</h2> : null}
      {children}
    </div>
  );
  if (!scrim) return sheet;
  return <div className="mmr-sheet-scrim">{sheet}</div>;
}
