import type { HTMLAttributes, ReactNode } from 'react';
import { useStyles } from '../styles/useStyles';
import type { Ground } from './types';

export interface ScreenProps extends HTMLAttributes<HTMLDivElement> {
  /** One ground per screen. `earning` (yellow) is the only ground that switches text to dark ink. */
  ground: Ground;
  children?: ReactNode;
}
/** 390x844 story-card canvas. Owns the ground colour, the insets and the text colour for everything inside. */
export function Screen({ ground, children, className = '', ...rest }: ScreenProps) {
  useStyles();
  return (
    <div className={`kmr-screen ${className}`} data-ground={ground} {...rest}>
      {children}
    </div>
  );
}
