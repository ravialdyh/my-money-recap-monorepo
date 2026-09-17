import type { ButtonHTMLAttributes, ReactNode } from 'react';
import { Icon } from '../icons';
import { ensureStyles } from '../styles/ensureStyles';
import { CardTitle } from './Card';
import { IconWell } from './DataRow';
import { Eyebrow, HelperText } from './Headline';
import type { InteractiveState } from './types';

export interface PromptCardProps extends Omit<ButtonHTMLAttributes<HTMLButtonElement>, 'children' | 'title'> {
  /** Small label above the title, e.g. "Step 3 of 4". `null` (the default) removes it. */
  eyebrow?: ReactNode | null;
  /** The prompt itself, e.g. "Set your monthly salary". Styled as a `CardTitle`. */
  title: ReactNode;
  /** Sentence under the title, styled as `HelperText`. `null` (the default) removes it. */
  helper?: ReactNode | null;
  state?: InteractiveState;
  /**
   * The measured height of the filled card this prompt stands in for (PRD §6.8, D54, D71), so nothing
   * moves when the Member fills it in. Omit it where there is nothing to stand in for and the card
   * hugs its content.
   */
  height?: number;
}

/**
 * Empty-state card: the whole card is one tap target, not the chevron (PRD §13, D50), and it holds the
 * filled card's height so the layout never jumps. The card *is* the `<button>`, so `Eyebrow`,
 * `CardTitle` and `HelperText` take their `span` form.
 */
export function PromptCard({
  height,
  eyebrow = null,
  title,
  helper = null,
  state = 'default',
  className = '',
  style,
  ...rest
}: PromptCardProps) {
  ensureStyles();
  return (
    <button
      type="button"
      className={`mmr-card mmr-prompt-card mmr-pressable ${className}`}
      data-state={state}
      style={height == null ? style : { ...style, height }}
      disabled={state === 'disabled' || rest.disabled}
      {...rest}
    >
      <span className="mmr-prompt-card-body">
        {eyebrow ? <Eyebrow as="span">{eyebrow}</Eyebrow> : null}
        <CardTitle as="span">{title}</CardTitle>
        {helper ? <HelperText as="span">{helper}</HelperText> : null}
      </span>
      <IconWell tone="violet" size={44} glyphSize={20}>
        <Icon name="chevron" />
      </IconWell>
    </button>
  );
}
