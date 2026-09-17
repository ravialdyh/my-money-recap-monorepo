import type { ButtonHTMLAttributes, ReactNode } from 'react';
import { ensureStyles } from '../styles/ensureStyles';
import type { InteractiveState } from './types';

export interface ButtonProps extends Omit<ButtonHTMLAttributes<HTMLButtonElement>, 'children'> {
  icon?: ReactNode;
  /** `primary` ink pill with a white label. `secondary` white pill with an ink label, for the second action in a pair. */
  variant?: 'primary' | 'secondary';
  state?: InteractiveState;
  children: ReactNode;
}
/** 52px full-width pill, optional 20px stroked icon. Ink by default; `secondary` is the white pill. */
export function Button({ icon, variant = 'primary', state = 'default', children, className = '', ...rest }: ButtonProps) {
  ensureStyles();
  return (
    <button type="button" className={`mmr-button mmr-pressable ${className}`} data-variant={variant} data-state={state} disabled={state === 'disabled' || rest.disabled} {...rest}>
      {icon}
      {children}
    </button>
  );
}
