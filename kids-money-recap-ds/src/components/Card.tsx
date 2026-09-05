import type { HTMLAttributes, ReactNode } from 'react';
import { useStyles } from '../styles/useStyles';

export interface CardProps extends HTMLAttributes<HTMLDivElement> {
  /** `surface` white (default) · `accent` violet with white text, for a second card on the yellow ground · `split` row layout with an illustration column. */
  variant?: 'surface' | 'accent' | 'split';
  /** Adds top padding so an Illustration above can hang into the card by `space.overlap`. */
  underIllustration?: boolean;
  /** Plays the entry animation (slide up 24px + fade). */
  animateIn?: boolean;
  children?: ReactNode;
}
/** 28px-radius card. Depth comes from colour and overlap, never from shadow. */
export function Card({ variant = 'surface', underIllustration = false, animateIn = false, children, className = '', ...rest }: CardProps) {
  useStyles();
  return (
    <div className={`kmr-card ${animateIn ? 'kmr-card-enter' : ''} ${className}`} data-variant={variant} data-under-illustration={underIllustration ? 'true' : undefined} {...rest}>
      {children}
    </div>
  );
}
export function CardTitle({ children, className = '', ...rest }: HTMLAttributes<HTMLHeadingElement>) {
  useStyles();
  return <h2 className={`kmr-card-title ${className}`} {...rest}>{children}</h2>;
}
