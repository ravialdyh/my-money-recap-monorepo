import type { ButtonHTMLAttributes, ReactNode } from 'react';
import { useStyles } from '../styles/useStyles';
import { Icon } from '../icons';
import { IconWell } from './DataRow';
import type { InteractiveState } from './types';

export interface PromptCardProps extends Omit<ButtonHTMLAttributes<HTMLButtonElement>, 'children' | 'title'> {
  /**
   * Which filled card this one stands in for, and therefore how tall it is (PRD §6.8, D54).
   * `hero` 191pt and `target` 136pt are the two Home cards exactly, so nothing on the screen moves
   * when the Member fills them in. `auto` hugs its content, for a Settings row.
   */
  size?: 'hero' | 'target' | 'auto';
  /** Small label above the title, e.g. "Step 3 of 4". `null` (the default) removes it. */
  eyebrow?: ReactNode | null;
  /** The prompt itself, e.g. "Set your monthly salary". Styled as a `CardTitle`. */
  title: ReactNode;
  /** Sentence under the title, styled as `HelperText`. `null` (the default) removes it. */
  helper?: ReactNode | null;
  state?: InteractiveState;
}
/**
 * Empty-state card: the whole card is one tap target, not the chevron (PRD §13, D50). It stands in
 * for a filled `Card` — the Home hero, the Home target card, the onboarding step 4 slider card, the
 * Recap screen 4 mascot, a Settings row — and holds that card's height so the layout never jumps.
 *
 * The card *is* the `<button>`, so the eyebrow, title and helper are spans carrying the `Eyebrow`,
 * `CardTitle` and `HelperText` classes rather than the components themselves: a `<button>` may not
 * contain an `<h2>` or a `<p>`. The drawn result is identical.
 */
export function PromptCard({
  size = 'auto', eyebrow = null, title, helper = null, state = 'default', className = '', ...rest
}: PromptCardProps) {
  useStyles();
  return (
    <button
      type="button" className={`mmr-card mmr-prompt-card mmr-pressable ${className}`}
      data-size={size} data-state={state}
      disabled={state === 'disabled' || rest.disabled} {...rest}
    >
      <span className="mmr-prompt-card-body">
        {eyebrow ? <span className="mmr-eyebrow" data-tone="secondary">{eyebrow}</span> : null}
        <span className="mmr-card-title">{title}</span>
        {helper ? <span className="mmr-helper">{helper}</span> : null}
      </span>
      <IconWell tone="violet"><Icon name="chevron" /></IconWell>
    </button>
  );
}
