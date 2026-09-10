import type { HTMLAttributes, ReactNode } from 'react';
import { useStyles } from '../styles/useStyles';

export interface InlineErrorProps extends HTMLAttributes<HTMLDivElement> {
  /**
   * Trailing control, e.g. a `Try again` link on the rate-fetch failure (SCREEN-STATES D6) or a
   * link to Profile on an undeliverable address (D9). `null` (the default) removes it.
   */
  action?: ReactNode | null;
  children: ReactNode;
}
/**
 * The app's only error surface (PRD D67 removed `Toast`): sign-in, email entry, a failed spend
 * write, a failed rate fetch, a failed export. It is a block in the flow — it pushes what follows
 * down and never overlays it, which is the whole reason it beat the toast (PRD §6.1, §6.19).
 */
export function InlineError({ action = null, children, className = '', ...rest }: InlineErrorProps) {
  useStyles();
  return (
    <div className={`mmr-inline-error ${className}`} role="alert" {...rest}>
      <span className="mmr-inline-error-dot" aria-hidden="true" />
      <span className="mmr-inline-error-body">{children}</span>
      {action ? <span className="mmr-inline-error-action">{action}</span> : null}
    </div>
  );
}
