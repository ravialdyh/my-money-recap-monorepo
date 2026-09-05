import type { HTMLAttributes } from 'react';
import { useStyles } from '../styles/useStyles';

export interface IllustrationProps extends HTMLAttributes<HTMLDivElement> {
  /** Transparent WebP/PNG at 2x the display width. Generate from the style block in illustration-prompts.md. */
  src: string;
  alt?: string;
  /** Display width in px. Observed: 310 cover, 200 lemonade, 168 vending, 340 squirrel. */
  width: number;
  /**
   * Cap on the rendered height. Set this wherever one slot shows several illustrations: the
   * artwork varies in aspect ratio, so a shared `width` alone lets a tall subject push the
   * content below it off the screen (the Hustler bee did exactly that on the archetype screen).
   */
  maxHeight?: number;
}
/** Flexible slot that hangs the illustration `space.overlap` into the card below it. */
export function Illustration({ src, alt = '', width, maxHeight, className = '', ...rest }: IllustrationProps) {
  useStyles();
  return (
    <div className={`mmr-illo ${className}`} {...rest}>
      <img src={src} alt={alt} style={{ width, maxHeight, objectFit: maxHeight ? 'contain' : undefined }} />
    </div>
  );
}
