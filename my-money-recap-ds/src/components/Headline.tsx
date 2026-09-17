import type { HTMLAttributes, ReactNode } from 'react';
import { ensureStyles } from '../styles/ensureStyles';

export interface HeadlineProps extends HTMLAttributes<HTMLHeadingElement> {
  /** `l` 48px cover headline · `m` 40px screen headline · `compact` 36px when three lines are unavoidable · `display` 68px archetype name. */
  size?: 'l' | 'm' | 'compact' | 'display';
  as?: 'h1' | 'h2';
  children: ReactNode;
}
/** Fredoka 700 headline. Inherits colour from the Screen so it is white on colour and ink on yellow. */
export function Headline({ size = 'm', as: Tag = 'h1', children, className = '', ...rest }: HeadlineProps) {
  ensureStyles();
  return (
    <Tag className={`mmr-headline ${className}`} data-size={size} {...rest}>
      {children}
    </Tag>
  );
}

export function SubCopy({ children, className = '', ...rest }: HTMLAttributes<HTMLParagraphElement>) {
  ensureStyles();
  return (
    <p className={`mmr-subcopy ${className}`} {...rest}>
      {children}
    </p>
  );
}
export interface HelperTextProps extends HTMLAttributes<HTMLElement> {
  /** `span` when the helper sits inside a `<button>`, which may not contain a `<p>`. Drawn identically. */
  as?: 'p' | 'span';
}
export function HelperText({ as: Tag = 'p', children, className = '', ...rest }: HelperTextProps) {
  ensureStyles();
  return (
    <Tag className={`mmr-helper ${className}`} {...rest}>
      {children}
    </Tag>
  );
}
export interface EyebrowProps extends HTMLAttributes<HTMLElement> {
  tone?: 'secondary' | 'coral';
  /** `span` when the eyebrow sits inside a `<button>`, which may not contain a `<div>`. Drawn identically. */
  as?: 'div' | 'span';
}
/** Small label above a stat or callout. `coral` is the uppercase callout style ("THE SPLURGE"). */
export function Eyebrow({ tone = 'secondary', as: Tag = 'div', children, className = '', ...rest }: EyebrowProps) {
  ensureStyles();
  return (
    <Tag className={`mmr-eyebrow ${className}`} data-tone={tone} {...rest}>
      {children}
    </Tag>
  );
}
