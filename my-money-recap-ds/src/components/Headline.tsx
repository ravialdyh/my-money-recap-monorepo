import type { HTMLAttributes, ReactNode } from 'react';
import { useStyles } from '../styles/useStyles';

export interface HeadlineProps extends HTMLAttributes<HTMLHeadingElement> {
  /** `l` 48px cover headline · `m` 40px screen headline · `compact` 36px when three lines are unavoidable · `display` 68px archetype name. */
  size?: 'l' | 'm' | 'compact' | 'display';
  as?: 'h1' | 'h2';
  children: ReactNode;
}
/** Fredoka 700 headline. Inherits colour from the Screen so it is white on colour and ink on yellow. */
export function Headline({ size = 'm', as: Tag = 'h1', children, className = '', ...rest }: HeadlineProps) {
  useStyles();
  return <Tag className={`mmr-headline ${className}`} data-size={size} {...rest}>{children}</Tag>;
}

export function SubCopy({ children, className = '', ...rest }: HTMLAttributes<HTMLParagraphElement>) {
  useStyles();
  return <p className={`mmr-subcopy ${className}`} {...rest}>{children}</p>;
}
export function HelperText({ children, className = '', ...rest }: HTMLAttributes<HTMLParagraphElement>) {
  useStyles();
  return <p className={`mmr-helper ${className}`} {...rest}>{children}</p>;
}
export interface EyebrowProps extends HTMLAttributes<HTMLDivElement> { tone?: 'secondary' | 'coral'; }
/** Small label above a stat or callout. `coral` is the uppercase callout style ("THE SPLURGE"). */
export function Eyebrow({ tone = 'secondary', children, className = '', ...rest }: EyebrowProps) {
  useStyles();
  return <div className={`mmr-eyebrow ${className}`} data-tone={tone} {...rest}>{children}</div>;
}
