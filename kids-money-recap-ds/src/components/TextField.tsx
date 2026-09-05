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
}
/** Single-line text input that lives inside a Card. Nunito 800 ink on the field surface, violet focus ring. */
export function TextField({ label, prefix, suffix, helper, state = 'default', className = '', ...rest }: TextFieldProps) {
  useStyles();
  return (
    <label className={`kmr-field-wrap ${className}`}>
      {label ? <span className="kmr-field-label">{label}</span> : null}
      <span className="kmr-field" data-state={state}>
        {prefix ? <span className="kmr-field-affix">{prefix}</span> : null}
        <input className="kmr-field-input" disabled={state === 'disabled' || rest.disabled} {...rest} />
        {suffix ? <span className="kmr-field-affix" data-side="end">{suffix}</span> : null}
      </span>
      {helper ? <span className="kmr-field-helper" data-state={state}>{helper}</span> : null}
    </label>
  );
}
