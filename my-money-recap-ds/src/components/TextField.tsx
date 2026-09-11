import type { InputHTMLAttributes, ReactNode } from 'react';
import { useStyles } from '../styles/useStyles';

/** Field states. `focus` and `invalid` draw a 2px ring; `disabled` follows the 0.4 opacity rule. */
export type FieldState = 'default' | 'focus' | 'invalid' | 'disabled';

export interface TextFieldProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'size' | 'prefix'> {
  /** Small ink label above the field. */
  label?: string;
  /** Fixed leading content, e.g. a currency symbol. Never editable. */
  prefix?: ReactNode;
  /** Fixed trailing content, e.g. a unit or a character count. */
  suffix?: ReactNode;
  /** Helper or error line under the field. Coral when `state` is `invalid`. */
  helper?: string;
  /** Forced state for previews. Real focus and `disabled` work without it. */
  state?: FieldState;
  /** Set by `AmountField` (PRD §11 step-down); lands on the wrapper so the affix and the input size together. */
  'data-step'?: string;
}
/** Single-line text input that lives inside a Card. Nunito 800 ink on the field surface, violet focus ring. */
export function TextField({ label, prefix, suffix, helper, state = 'default', className = '', 'data-step': step, ...rest }: TextFieldProps) {
  useStyles();
  return (
    <label className={`mmr-field-wrap ${className}`} data-step={step}>
      {label ? <span className="mmr-field-label">{label}</span> : null}
      <span className="mmr-field" data-state={state}>
        {prefix ? <span className="mmr-field-affix">{prefix}</span> : null}
        <input className="mmr-field-input" disabled={state === 'disabled' || rest.disabled} {...rest} />
        {suffix ? <span className="mmr-field-affix" data-side="end">{suffix}</span> : null}
      </span>
      {helper ? <span className="mmr-field-helper" data-state={state}>{helper}</span> : null}
    </label>
  );
}
